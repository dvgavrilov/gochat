package messaging

import (
	"encoding/json"
	"errors"
	"fmt"
	"leto-yanao-1/service/source/models"
	"leto-yanao-1/service/source/persistence"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// ErrAddMessage error
	ErrAddMessage = errors.New("error while creating a new message")
	// ErrMarschalingMessage error
	ErrMarschalingMessage = errors.New("error marshaling message")
	// ErrConversationNotFound error
	ErrConversationNotFound = errors.New("conversation not found")
	// ErrGettingConversation error
	ErrGettingConversation = errors.New("error while getting a conversation")
	// ErrGetMessages error
	ErrGetMessages = errors.New("error while getting messages by conversation id")
	// ErrMarshalingResponse error
	ErrMarshalingResponse = errors.New("error marshaling result from a server")
	// ErrGetMessage error
	ErrGetMessage = errors.New("error while getting a message")
	// ErrBadEventArgs error
	ErrBadEventArgs = errors.New("received event arguments invalid")
	// ErrReadMessages error
	ErrReadMessages = errors.New("error while reading user messages")
	// ErrUpdateMessage error
	ErrUpdateMessage = errors.New("error while updating a message")
	// ErrChatRoomNotFound error
	ErrChatRoomNotFound = errors.New("error finding a chat room, you shoud join first")
)

const (
	// GetMessageListEvent error
	GetMessageListEvent = "Event.GetMessageList"
	// SendMessageEvent error
	SendMessageEvent = "Event.SendMessage"
	// ReceiveMessageEvent error
	ReceiveMessageEvent = "Event.ReceiveMessage"
	// ReadMessageEvent error
	ReadMessageEvent = "Event.ReadMessage"
	// GetUnreadInfoEvent const
	GetUnreadInfoEvent = "Event.GetUnreadInfo"
	// GetUnreadMessagesEvent const
	GetUnreadMessagesEvent = "Event.GetUnreadMessages"
)

type getMessageListEventArgs struct {
	SessionChannel string `json:"session_channel"`
	ExecutorID     uint   `json:"executor_id"`
}

type getMessageListEventResult struct {
	Messages []*message
}

type sendMessageEventArgs struct {
	SessionChannel string `json:"session_channel"`
	Content        string `json:"content"`
	ContentType    uint   `json:"content_type"` // 1 or 2. 1 is a text, and 2 is an image. If no content type, we treat it as text.
	SenderID       uint   `json:"sender_id"`
}

type sendMessageEventResult struct {
	Message *message `json:"message"`
}

type receiveMessageEventResult struct {
	Message *message `json:"message"`
}

type readMessageEventArgs struct {
	ExecutorID uint `json:"executor_id"`
	MessageID  uint `json:"message_id"`
}

type readMessageEventResult struct {
	ExecutorID uint `json:"executor_id"`
	MessageID  uint `json:"message_id"`
}

type message struct {
	ID             uint      `json:"id"`
	SessionChannel string    `json:"session_channel"`
	Content        string    `json:"content"`
	ContentType    uint      `json:"content_type"`
	SenderID       uint      `json:"sender_id"`
	Read           bool      `json:"read"`
	CreatedAt      time.Time `json:"create_at"`
	UpdateAt       time.Time `json:"update_at"`
}

type getUnreadInfoArgs struct {
	UserID uint `json:"user_id"`
}

type getUnreadInfoResult struct {
	UserID      uint `json:"user_id"`
	UnreadCount int  `json:"unread_count"`
}

type getUnreadMessagesEventArgs struct {
	ExecutorID uint `json:"executor_id"`
}

type getUnreadMessagesEventResult struct {
	Messages []*message
}

func onGetUnreadMessagesList(e *Event, c *Chatter) (*EventResult, error) {
	if !c.IsModerator {
		return e.getErrorResponse(ErrUnauthorized), nil
	}

	args := &getUnreadMessagesEventArgs{}
	err := convertFromRaw([]byte(reflect.ValueOf(e.Args).String()), args)
	if err != nil {
		logrus.Error(err)
		return e.getErrorResponse(ErrMarschalingMessage), nil
	}

	if args.ExecutorID != c.UserID {
		return e.getErrorResponse(ErrUnauthorized), nil
	}

	messages, err := persistence.GetMessagesProvider().GetUnreadMessages(c.UserID)
	if err != nil {
		logrus.Error(err)
		return e.getErrorResponse(ErrReadMessages), nil
	}

	return &EventResult{
		Name: (*e).Name,
		Ok:   true,
		Result: getMessageListEventResult{
			Messages: convertMessages(messages),
		},
	}, nil
}

func onGetMessageList(e *Event, c *Chatter) (*EventResult, error) {
	args := &getMessageListEventArgs{}
	err := convertFromRaw([]byte(reflect.ValueOf(e.Args).String()), args)
	if err != nil {
		logrus.Error(err)
		return e.getErrorResponse(ErrMarschalingMessage), nil
	}

	sc, err := ParseSessionChannel(args.SessionChannel)
	if err != nil {
		logrus.Error(err)
		return e.getErrorResponse(err), nil
	}

	logrus.Info(fmt.Sprintf("Received a new %v event with following parameters: SessionChannel:%s", e.Name, sc.ToString()))

	if args.ExecutorID != c.UserID {
		return e.getErrorResponse(ErrUnauthorized), nil
	}

	conversation, err := getConversation(sc)
	if err != nil {
		return e.getErrorResponse(ErrConversationNotFound), nil
	}

	messages, err := persistence.GetMessagesProvider().GetMessages(conversation.ID)
	if err != nil {
		logrus.Error(err)
		return nil, ErrGetMessages
	}

	for i := range *messages {
		(*messages)[i].Read = isMessageRead(&((*messages)[i]), c.UserID)
	}

	return &EventResult{
		Name: (*e).Name,
		Ok:   true,
		Result: getMessageListEventResult{
			Messages: convertMessages(messages),
		},
	}, nil
}

func onSendMessage(e *Event, c *Chatter) (*EventResult, error) {
	args := &sendMessageEventArgs{}
	err := convertFromRaw([]byte(reflect.ValueOf(e.Args).String()), args)
	if err != nil {
		logrus.Error(err)
		return e.getErrorResponse(ErrMarschalingMessage), nil
	}

	if c.UserID != args.SenderID {
		return e.getErrorResponse(ErrUnauthorized), nil
	}

	sc, err := ParseSessionChannel(args.SessionChannel)
	if err != nil {
		logrus.Error(err)
		return e.getErrorResponse(err), nil
	}

	logrus.Info(fmt.Sprintf("Received a new %v event with following parameters: SessionChannel:%s", e.Name, sc.ToString()))

	conversation, err := getConversation(sc)
	if err != nil {
		return e.getErrorResponse(ErrConversationNotFound), nil
	}

	if conversation == nil {
		return e.getErrorResponse(ErrConversationNotFound), nil
	}

	msg := &models.Message{
		SenderID:       args.SenderID,
		Content:        args.Content,
		ConversationID: conversation.ID,
		ApplicationID:  conversation.ApplicationID,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if args.ContentType == 0 || args.ContentType > 2 {
		msg.ContentType = models.ContentText
	} else {
		msg.ContentType = args.ContentType
	}

	msg, err = persistence.GetMessagesProvider().AddMessage(msg)
	if err != nil {
		logrus.Error(err)
		return nil, ErrAddMessage
	}

	// A message was added, now, we need to add an unread information, so,
	// all participants will be able to mark that the message was read by him
	if len(conversation.Participants) > 0 {
		for _, p := range conversation.Participants {
			if p.UserID == msg.SenderID {
				continue
			}

			insertUnreadInfo(msg, p.UserID)
		}
	}

	res := convertMessage(msg)
	receiveMessage := &EventResult{
		Name: ReceiveMessageEvent,
		Ok:   true,
		Result: receiveMessageEventResult{
			Message: res,
		},
	}

	raw, err := json.Marshal(receiveMessage)
	if err != nil {
		logrus.Error(err)
		return e.getErrorResponse(ErrMarshalingResponse), nil
	}

	key := fmt.Sprintf("%v", sc.ApplicationID)
	if room, ok := c.Rooms[key]; ok {
		err = room.broadcast(
			raw,
			func(toCheck *Chatter) bool { return c != toCheck })

		if err == ErrNoChatterMatch {
			err = hub.broadcast(raw,
				func(toCheck *Chatter) bool {
					return c != toCheck && toCheck.IsModerator
				})

			insertUnreadInfoGlobal(msg)
		}
	} else {
		return e.getErrorResponse(ErrChatRoomNotFound), nil
	}

	return &EventResult{
		Name: (*e).Name,
		Ok:   true,
		Result: sendMessageEventResult{
			Message: res,
		},
	}, nil
}

func onReadMessage(e *Event, c *Chatter) (*EventResult, error) {

	args := &readMessageEventArgs{}
	err := convertFromRaw([]byte(reflect.ValueOf(e.Args).String()), args)
	if err != nil {
		logrus.Error(err)
		return e.getErrorResponse(ErrMarschalingMessage), nil
	}

	logrus.Info(fmt.Sprintf("Received a new %v event", e.Name))

	if c.UserID != args.ExecutorID {
		return e.getErrorResponse(ErrUnauthorized), nil
	}

	if c.IsModerator {
		err = persistence.GetUnreadInfoManager().MarkAsRead(args.MessageID, 0)
		if err != nil {
			return e.getErrorResponse(ErrUpdateMessage), nil
		}
	}

	err = persistence.GetUnreadInfoManager().MarkAsRead(args.MessageID, c.UserID)
	if err != nil {
		return e.getErrorResponse(ErrUpdateMessage), nil
	}

	return &EventResult{
		Name: (*e).Name,
		Ok:   true,
		Result: readMessageEventResult{
			ExecutorID: args.ExecutorID,
			MessageID:  args.MessageID,
		},
	}, nil
}

func onGetUnreadInfo(e *Event, c *Chatter) (*EventResult, error) {

	args := &getUnreadInfoArgs{}
	err := convertFromRaw([]byte(reflect.ValueOf(e.Args).String()), args)
	if err != nil {
		logrus.Error(err)
		return e.getErrorResponse(ErrMarschalingMessage), nil
	}

	if c.UserID != args.UserID {
		return e.getErrorResponse(ErrUnauthorized), nil
	}

	logrus.Info(fmt.Sprintf("Received a new %v event with following parameters: UserID:%v", e.Name, args.UserID))

	var count int
	if c.IsModerator {
		count, err = persistence.GetUnreadInfoManager().GetForUserAndGlobal(args.UserID)

	} else {
		count, err = persistence.GetUnreadInfoManager().GetForUser(args.UserID)
	}

	if err != nil {
		return e.getErrorResponse(ErrReadMessages), nil
	}

	return &EventResult{
		Name: (*e).Name,
		Ok:   true,
		Result: getUnreadInfoResult{
			UserID:      args.UserID,
			UnreadCount: count,
		},
	}, nil

}

func convertMessages(messages *[]models.Message) []*message {

	ret := make([]*message, 0)
	if messages == nil {
		return ret
	}

	for _, c := range *messages {
		ret = append(ret, convertMessage(&c))
	}

	return ret
}

func convertMessage(model *models.Message) *message {
	return &message{
		ID:          model.ID,
		SenderID:    model.SenderID,
		Content:     model.Content,
		ContentType: model.ContentType,
		Read:        model.Read,
		CreatedAt:   model.CreatedAt,
		UpdateAt:    model.UpdatedAt,
		SessionChannel: SessionChannel{
			ApplicationID: model.ApplicationID,
		}.ToString(),
	}
}

func insertUnreadInfoGlobal(msg *models.Message) error {
	uinfo := &models.UnreadInfo{
		MessageID:      msg.ID,
		ConversationID: msg.ConversationID,
	}

	return persistence.GetUnreadInfoManager().Add(uinfo)
}

func insertUnreadInfos(msg *models.Message, participants ...uint) error {
	var err error
	for _, p := range participants {
		err = insertUnreadInfo(msg, p)
		if err != nil {
			break
		}
	}

	return err
}

func insertUnreadInfo(msg *models.Message, participantID uint) error {
	uinfo := &models.UnreadInfo{
		MessageID:      msg.ID,
		ConversationID: msg.ConversationID,
		ParticipantID:  participantID,
	}

	return persistence.GetUnreadInfoManager().Add(uinfo)
}

func isMessageRead(msg *models.Message, userID uint) bool {
	for _, ui := range msg.UnreadInfo {

		if ui.MessageID != msg.ID {
			continue
		}

		if ui.ConversationID != msg.ConversationID {
			continue
		}

		return ui.Read
	}

	return false
}
