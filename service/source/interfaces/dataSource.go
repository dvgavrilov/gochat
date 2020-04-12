package interfaces

import (
	"io"
	"leto-yanao-1/service/source/models"
)

type DataSource interface {
	Init() error
	io.Closer
	MessagesProvider
	ConversationsProvider
}

// MessagesProvider struct
type MessagesProvider interface {
	GetMessages(conversationID uint) (*[]models.Message, error)
	GetUnreadMessages(userID uint) (*[]models.Message, error)
	AddMessage(message *models.Message) (*models.Message, error)
}

// ConversationsProvider структура
type ConversationsProvider interface {
	GetByUserID(userID uint) (*[]models.Conversation, error)
	GetByApplicationID(applicationID uint) (*models.Conversation, error)
	Add(conversation *models.Conversation) (*models.Conversation, error)
}

// ParticipantsProvider struct
type ParticipantsProvider interface {
	Add(participant *models.Participant) error
	GetByConversationID(conversationID uint) (*[]models.Participant, error)
}

// UnreadInfoManager struct
type UnreadInfoManager interface {
	Add(unreadInfo *models.UnreadInfo) error
	GetByMessageID(messageID uint) (*[]models.UnreadInfo, error)
	GetForUser(participantID uint) (int, error)
	GetForUserAndGlobal(participantID uint) (int, error)
	MarkAsRead(messageID uint, participantID uint) error
}
