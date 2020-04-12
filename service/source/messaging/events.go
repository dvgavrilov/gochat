package messaging

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	// ErrSessionChannelInvalid error
	ErrSessionChannelInvalid = errors.New("received session channel value is invalid")
)

// Event structure
type Event struct {
	Name string      `json:"name"`
	Args interface{} `json:"args"`
}

// EventResult structure
type EventResult struct {
	Name   string      `json:"name"`
	Ok     bool        `json:"ok"`
	Result interface{} `json:"result"`
}

// EventHandler function - обработчик сообщения
type EventHandler func(*Event, *Chatter) (*EventResult, error)

// SessionChannel - структура
type SessionChannel struct {
	ApplicationID uint
	// CustomerID    uint
	// UserID        uint
}

func (e *Event) getErrorResponse(msg error) *EventResult {
	return &EventResult{
		Name:   (*e).Name,
		Ok:     false,
		Result: msg.Error(),
	}
}

// ToString func
func (sc SessionChannel) ToString() string {
	return fmt.Sprintf("%v", sc.ApplicationID)
}

// ParseSessionChannel - функция отвечающая за чтение строкового представления сессионого ключа,
// который имеет вид customerid_userid, где customerid это уникальный идентификатор клиента-родителя,
// а userid уникальный идентификатор модератора соответственно
func ParseSessionChannel(value string) (*SessionChannel, error) {
	if value == "" {
		return nil, ErrSessionChannelInvalid
	}

	parts := strings.Split(value, "_")
	if len(parts) != 1 {
		return nil, ErrSessionChannelInvalid
	}

	arg1, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return nil, ErrSessionChannelInvalid
	}
	// arg2, err := strconv.ParseUint(parts[1], 10, 32)
	// if err != nil {
	// 	return nil, ErrSessionChannelInvalid
	// }
	// arg3, err := strconv.ParseUint(parts[2], 10, 32)
	// if err != nil {
	// 	return nil, ErrSessionChannelInvalid
	// }

	return &SessionChannel{
		ApplicationID: uint(arg1),
	}, nil
}

func rawToEvent(bytes []byte) (*Event, error) {
	event := &Event{}
	err := json.Unmarshal(bytes, event)
	return event, err
}

func convertFromRaw(bytes []byte, out interface{}) error {

	err := json.Unmarshal(bytes, out)
	return err
}
