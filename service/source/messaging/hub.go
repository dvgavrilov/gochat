package messaging

import (
	"errors"
	"fmt"

	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/dvgavrilov/gochat/service/source/config"
	"github.com/dvgavrilov/gochat/service/source/customerrors"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var (
	//ErrApplicationIDInvalid error
	ErrApplicationIDInvalid = errors.New("request id parameter is missing or invalid")
	// ErrRoomLimitExceed error
	ErrRoomLimitExceed = errors.New("room participants exceeded")
	// ErrChatterDuplicate error
	ErrChatterDuplicate = errors.New("a user with that name is already registered")
	// ErrUnknownChatter error
	ErrUnknownChatter = errors.New("a user is not registered")
	// ErrLimitChattersExceed error
	ErrLimitChattersExceed = errors.New("amount of chatters exceed")
	// ErrTokenBadStructure error
	ErrTokenBadStructure = errors.New("received token has wrong structure")
	// ErrNoChatterMatch error
	ErrNoChatterMatch = errors.New("error finding the chatters to broadcast the message")
)

const (
	requestID = "request_id"
	adminID   = "admin_id"
)

var hub *Hub

// Hub структура
type Hub struct {
	mux      sync.Mutex
	Rooms    map[string]*chatRoom
	Chatters map[*Chatter]uint
}

type predicate func(*Chatter) bool

type chatterCounter struct {
	chatter *Chatter
	count   uint
}

type chatRoom struct {
	ID       string
	RoomID   uint
	Chatters map[*Chatter]bool
}

// WebSocket структура
type WebSocket struct {
	Conn *websocket.Conn
}

// Init Функция
func Init(r *chi.Mux) {
	r.Use(verifier(jwtauth.New("HS256", jwtSecret, nil)))
	r.Use(authorize)

	registerWsListener(r)
}

var upgrader = websocket.Upgrader{

	ReadBufferSize:  config.MainConfiguration.WebSocketSettings.ReadBufferSize,
	WriteBufferSize: config.MainConfiguration.WebSocketSettings.WriteBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		if config.MainConfiguration.WebSocketSettings.Origin == "" {
			return true
		}

		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			return false
		}

		if strings.ToLower(host) == strings.ToLower(config.MainConfiguration.WebSocketSettings.Origin) {
			return true
		}
		return false
	},
}

func registerWsListener(r *chi.Mux) {
	hub = &Hub{
		Chatters: make(map[*Chatter]uint),
		Rooms:    make(map[string]*chatRoom),
	}

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleRequests(w, r)
	})
}

func handleRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sid, err := getSenderID(r)
	if err != nil {
		logrus.Error(err)
		http.Error(w, "the sid parameter is invalid", http.StatusBadRequest)
		return
	}

	// if _, ok := hub.Chatters[sid]; ok {
	// 	logrus.Error(ErrChatterDuplicate)
	// 	http.Error(w, "the sid parameter is invalid", http.StatusBadRequest)
	// 	return
	// }

	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims == nil {
		http.Error(w, ErrTokenBadStructure.Error(), http.StatusBadRequest)
		return
	}

	var iscustomer bool
	if _, ok := claims[requestID]; ok {
		iscustomer = true
	}

	var isadmin bool
	if _, ok := claims[adminID]; ok {
		isadmin = true
	}

	if iscustomer && isadmin {
		logrus.Error(fmt.Sprintf("token contains both parameters: a request_id and an admin_id"))
		http.Error(w, ErrTokenBadStructure.Error(), http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, getUpgraderWebSocketHeader())
	if err != nil {
		logrus.Error(err)
		http.Error(w, "unable to upgrade", http.StatusInternalServerError)
		return
	}

	chatter := &Chatter{
		UserID:      sid,
		IsCustomer:  iscustomer,
		IsModerator: isadmin,
		WebSocket:   &WebSocket{Conn: conn},
		Rooms:       make(map[string]*chatRoom),
		Events:      make(map[string]EventHandler),
		Out:         make(chan []byte),
	}

	chatter.On(GetConversationListEvent, onGetConversationList)
	chatter.On(AddConversationEvent, onAddConversation)
	chatter.On(GetMessageListEvent, onGetMessageList)
	chatter.On(SendMessageEvent, onSendMessage)
	chatter.On(ReadMessageEvent, onReadMessage)
	chatter.On(GetUnreadInfoEvent, onGetUnreadInfo)
	chatter.On(GetUnreadMessagesEvent, onGetUnreadMessagesList)

	hub.register(chatter)

	logrus.Info(fmt.Sprintf("A chatter with id %v registered and created a socket for it.", sid))

	go chatter.Reader()
	go chatter.Writer()
}

func getSenderID(r *http.Request) (uint, error) {

	sid, err := strconv.ParseUint(r.FormValue("sid"), 10, 32)
	if err != nil {
		logrus.Error(err)
		return 0, ErrApplicationIDInvalid
	}

	return uint(sid), nil
}

func (h *Hub) register(chatter *Chatter) error {
	h.mux.Lock()
	defer h.mux.Unlock()

	if chatter == nil {
		return customerrors.ErrArgumentNilError
	}

	// if _, ok := h.Chatters[chatter.UserID]; ok {
	// 	return ErrChatterDuplicate
	// }
	h.Chatters[chatter] = chatter.UserID

	logrus.Info(fmt.Sprintf("A new chatter with id %v was registered", chatter.UserID))

	return nil
}

func (h *Hub) unregister(chatter *Chatter) error {
	h.mux.Lock()
	defer h.mux.Unlock()

	if _, ok := h.Chatters[chatter]; ok {
		delete(h.Chatters, chatter)
	}

	return nil
}

func (h *Hub) join(chatter *Chatter, chatID string) (*chatRoom, error) {

	h.mux.Lock()
	defer h.mux.Unlock()

	if h == nil || chatter == nil {
		return nil, customerrors.ErrArgumentNilError
	}

	if _, ok := h.Chatters[chatter]; !ok {
		return nil, ErrUnknownChatter
	}

	var ch *chatRoom
	var ok bool
	if ch, ok = h.Rooms[chatID]; !ok {
		ch = &chatRoom{
			ID:       chatID,
			Chatters: make(map[*Chatter]bool),
		}
		h.Rooms[chatID] = ch
	}

	if len(h.Rooms[chatID].Chatters) == config.MainConfiguration.ChatRoomSettings.MaxValueOfChatters {
		logrus.Info(fmt.Sprintf("The limit of chatters exceeds in %s chat room", chatID))
		return nil, ErrLimitChattersExceed
	}

	if _, ok = h.Rooms[chatID].Chatters[chatter]; !ok {
		h.Rooms[chatID].Chatters[chatter] = true
		logrus.Info(fmt.Sprintf("Chatter with id %v joined the %s chat room", chatter.UserID, chatID))
	}

	return ch, nil
}

func (h *Hub) leave(chatter *Chatter) error {
	h.mux.Lock()
	defer h.mux.Unlock()

	if len(chatter.Rooms) > 0 {
		for key := range chatter.Rooms {
			if room, ok := h.Rooms[key]; ok {
				if _, ok = room.Chatters[chatter]; ok {
					delete(h.Rooms[key].Chatters, chatter)
					logrus.Info(fmt.Sprintf("Chatter with id %v left the %s chat room", chatter.UserID, key))
					if len(room.Chatters) == 0 {
						delete(h.Rooms, key)
					}
				}
			}
		}
	}

	return nil
}

func (h *Hub) broadcast(message []byte, fn predicate) error {

	if h == nil {
		log.Panic("receiver is null")
	}

	atleastonce := false
	for k := range h.Chatters {
		if fn(k) {
			k.Out <- message
			atleastonce = true
		}
	}

	if atleastonce {
		return nil
	}
	return ErrNoChatterMatch
}

func (cr *chatRoom) broadcast(message []byte, fn predicate) error {
	if cr == nil {
		log.Panic("receiver is null")
	}

	atleastonce := false
	for k := range cr.Chatters {
		if fn(k) {
			k.Out <- message
			atleastonce = true
		}
	}

	if atleastonce {
		return nil
	}
	return ErrNoChatterMatch
}
