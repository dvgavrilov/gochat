package messaging

import (
	"encoding/json"
	"fmt"
	"leto-yanao-1/service/source/config"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

// Chatter structure
type Chatter struct {
	UserID      uint
	IsCustomer  bool
	IsModerator bool
	Rooms       map[string]*chatRoom
	Events      map[string]EventHandler
	WebSocket   *WebSocket
	Out         chan []byte
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Reader func
func (c *Chatter) Reader() {
	defer func() {
		hub.leave(c)
		c.WebSocket.Conn.Close()
		hub.unregister(c)
	}()

	c.WebSocket.Conn.SetReadLimit(int64(config.MainConfiguration.WebSocketSettings.ReadBufferSize))
	c.WebSocket.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.WebSocket.Conn.SetPongHandler(func(data string) error {
		c.WebSocket.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, m, err := c.WebSocket.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				logrus.Error("read message error", err)
			}
			break
		}

		e, err := rawToEvent(m)
		if err != nil {
			em := fmt.Sprintf("message parsing error: %v", err)
			logrus.Error(em)
			c.WebSocket.Conn.WriteJSON(em)
			// TODO - should we continue or break?
			continue
		}

		action, ok := c.Events[e.Name]
		if !ok {
			em := fmt.Sprintf("not supported event: %v", e.Name)
			logrus.Error(em)
			c.WebSocket.Conn.WriteJSON(em)
			// TODO - should we continue or break?
			continue
		}

		ret, err := action(e, c)
		if err != nil {
			em := fmt.Sprintf("a critical error happened, closing the socket.")
			logrus.Error(em)
			break
		}

		rm, err := json.Marshal(ret)
		if err != nil {
			em := fmt.Sprintf("event result serialization error: %v", err)
			logrus.Error(em)
			c.WebSocket.Conn.WriteJSON(em)
			// TODO - shold we continue or break?
			continue
		}

		c.Out <- rm
	}
}

// Writer func
func (c *Chatter) Writer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		hub.leave(c)
		ticker.Stop()
		hub.unregister(c)
		c.WebSocket.Conn.Close()
	}()

	for {
		select {
		case m, ok := <-c.Out:
			{
				c.WebSocket.Conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					c.WebSocket.Conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				w, err := c.WebSocket.Conn.NextWriter(websocket.TextMessage)
				if err != nil {
					em := fmt.Sprintf("getting next writer error: %v", err)
					logrus.Error(em)
					c.WebSocket.Conn.WriteJSON(em)
					return
				}

				_, err = w.Write(m)
				if err != nil {
					em := fmt.Sprintf("writing message back error: %v", err)
					logrus.Error(em)
					c.WebSocket.Conn.WriteJSON(em)
					return
				}

				err = w.Close()
				if err != nil {
					em := fmt.Sprintf("closing web socket writer error: %v", err)
					logrus.Error(em)
					c.WebSocket.Conn.WriteJSON(em)
					return
				}
			}
		case <-ticker.C:
			{
				err := c.WebSocket.wpong()
				if err != nil {
					return
				}
			}
		}
	}
}

func (w *WebSocket) wpong() error {
	w.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	err := w.write(websocket.PingMessage, []byte{})
	return err
}

func (w *WebSocket) write(messageType int, payload []byte) error {
	w.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	return w.Conn.WriteMessage(messageType, payload)
}

// On func
func (c *Chatter) On(eventName string, action EventHandler) *Chatter {
	(*c).Events[eventName] = action
	return c
}
