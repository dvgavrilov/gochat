package messaging

import (
	"errors"
	"fmt"
	"leto-yanao-1/service/source/models"
	"leto-yanao-1/service/source/persistence"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// ErrGetConnections error
	ErrGetConnections = errors.New("error while getting list of conversations")
	// ErrAddConversation error
	ErrAddConversation = errors.New("error while creating a new conversation")
	// ErrJoiningChat error
	ErrJoiningChat = errors.New("error while joining a chat room")
	// ErrUnauthorized error
	ErrUnauthorized = errors.New("a registered user is different from one you are trying to use")
)

const (
	// AddConversationEvent const
	AddConversationEvent = "Event.AddConversation"
	// GetConversationListEvent const
	GetConversationListEvent = "Event.GetConversationList"
)

type addConversationEventArgs struct {
	UserID        uint `json:"user_id"`
	ApplicationID uint `json:"application_id"`
}

type addConversationEventResult struct {
	Conversation conversation `json:"conversation"`
}

type getConversationListArgs struct {
	UserID uint `json:"user_id"`
}

type getConversationListResult struct {
	Conversations []*conversation `json:"conversations"`
}

type conversation struct {
	ID             uint      `json:"id"`
	SessionChannel string    `json:"session_channel"`
	ApplicationID  uint      `json:"application_id"`
	CreatedAt      time.Time `json:"create_at"`
	UpdateAt       time.Time `json:"update_at"`
}

func getConversation(sc *SessionChannel) (*models.Conversation, error) {
	return persistence.GetConversationsProvider().GetByApplicationID(sc.ApplicationID)
}

func onAddConversation(e *Event, c *Chatter) (*EventResult, error) {

	args := &addConversationEventArgs{}
	err := convertFromRaw([]byte(reflect.ValueOf(e.Args).String()), args)
	if err != nil {
		logrus.Error(err)
		return e.getErrorResponse(ErrMarschalingMessage), nil
	}

	if args.UserID != c.UserID {
		return e.getErrorResponse(ErrUnauthorized), nil
	}

	sc := SessionChannel{
		ApplicationID: args.ApplicationID,
	}

	logrus.Info(fmt.Sprintf("Received a new %v event with following parameters: SessionChannel:%s", e.Name, sc.ToString()))

	conv, err := getConversation(&sc)
	if err != nil {
		return e.getErrorResponse(ErrGettingConversation), nil
	}

	if conv == nil {
		newConversation := &models.Conversation{
			ApplicationID: args.ApplicationID,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Participants: []models.Participant{
				models.Participant{
					UserID: c.UserID,
				},
			},
		}

		newConversation, err = persistence.GetConversationsProvider().Add(newConversation)
		if err != nil {
			logrus.Error(err)
			return nil, ErrAddConversation
		}

		conv = newConversation
	} else {
		found := false
		for _, p := range conv.Participants {
			if p.UserID == c.UserID {
				found = true
				break
			}
		}

		if !found {
			p := &models.Participant{
				ConversationID: conv.ID,
				UserID:         c.UserID,
			}

			err = persistence.GetParticipantsProvider().Add(p)
			if err != nil {
				return e.getErrorResponse(ErrJoiningChat), nil
			}
		}
	}

	key := fmt.Sprintf("%v", args.ApplicationID)
	cr, err := hub.join(c, key)
	if err != nil {
		logrus.Error(err)
		return e.getErrorResponse(ErrJoiningChat), nil
	}
	if _, ok := c.Rooms[key]; !ok {
		c.Rooms[key] = cr
	}

	return &EventResult{
		Name: (*e).Name,
		Ok:   true,
		Result: &conversation{
			ID:             conv.ID,
			ApplicationID:  conv.ApplicationID,
			CreatedAt:      conv.CreatedAt,
			UpdateAt:       conv.UpdatedAt,
			SessionChannel: sc.ToString(),
		},
	}, nil
}

func onGetConversationList(e *Event, c *Chatter) (*EventResult, error) {

	args := &getConversationListArgs{}
	err := convertFromRaw([]byte(reflect.ValueOf(e.Args).String()), args)
	if err != nil {
		logrus.Error(err)
		return e.getErrorResponse(ErrMarschalingMessage), nil
	}

	logrus.Info(fmt.Sprintf("Received a new %v event with following parameters: UserID:%v", e.Name, args.UserID))

	if c.UserID != args.UserID {
		return e.getErrorResponse(ErrUnauthorized), nil
	}

	conversations, err := persistence.GetConversationsProvider().GetByUserID(args.UserID)
	if err != nil {
		logrus.Error(err)
		return nil, ErrGetConnections
	}

	return &EventResult{
		Name: (*e).Name,
		Ok:   true,
		Result: getConversationListResult{
			Conversations: convertConversations(conversations),
		},
	}, nil
}

func convertConversations(conversations *[]models.Conversation) []*conversation {

	ret := make([]*conversation, 0)
	if conversations == nil {
		return ret
	}

	for _, c := range *conversations {
		ret = append(ret, convertConversation(&c))
	}

	return ret
}

func convertConversation(model *models.Conversation) *conversation {

	return &conversation{
		ID:            model.ID,
		ApplicationID: model.ApplicationID,
		CreatedAt:     model.CreatedAt,
		UpdateAt:      model.UpdatedAt,
		SessionChannel: SessionChannel{
			ApplicationID: model.ApplicationID}.ToString(),
	}
}
