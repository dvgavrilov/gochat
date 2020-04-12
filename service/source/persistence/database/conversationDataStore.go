package database

import (
	"leto-yanao-1/service/source/interfaces"
	"leto-yanao-1/service/source/models"

	"github.com/jinzhu/gorm"
)

type conversationsManager struct {
	conversationStore conversationDataStore
	participantsStore participantDataStore
}

func (r *conversationsManager) Init(connection *Connection) {
	r.conversationStore = conversationDataStore{connection: connection}
	r.participantsStore = participantDataStore{connection: connection}
}

type conversationDataStore struct {
	connection *Connection
	interfaces.ConversationsProvider
}

func (r conversationsManager) GetByApplicationID(applicationID uint) (*models.Conversation, error) {
	ret, err := r.conversationStore.GetByApplicationID(applicationID)
	if err != nil {
		return nil, err
	}

	if ret == nil {
		return nil, nil
	}

	participants, err := r.participantsStore.GetByConversationID(ret.ID)
	if err != nil {
		return nil, err
	}

	ret.Participants = *participants
	return ret, nil
}

func (r conversationsManager) Add(conversation *models.Conversation) (*models.Conversation, error) {
	return r.conversationStore.Add(conversation)
}

func (r conversationsManager) GetByUserID(userID uint) (*[]models.Conversation, error) {
	ret, err := r.conversationStore.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	for i := range *ret {
		participants, err := r.participantsStore.GetByConversationID((*ret)[i].ID)
		if err != nil {
			return nil, err
		}

		(*ret)[i].Participants = *participants
	}

	return ret, nil
}

// GetByApplicationID func
func (r conversationDataStore) GetByApplicationID(applicationID uint) (*models.Conversation, error) {
	obj := &models.Conversation{}
	err := r.connection.db.Where("application_id = ?", applicationID).First(obj).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return obj, nil
}

// AddConversation func
func (r conversationDataStore) Add(conversation *models.Conversation) (*models.Conversation, error) {
	r.connection.db.NewRecord(conversation)
	err := r.connection.db.Create(&conversation).Error
	if err != nil {
		return nil, err
	}
	return conversation, err
}

func (r conversationDataStore) GetByUserID(userID uint) (*[]models.Conversation, error) {
	obj := &[]models.Conversation{}
	err := r.connection.db.Model(models.Conversation{}).
		Select("conversations.id, conversations.application_id, conversations.created_at, conversations.updated_at").
		Joins("join participants p on p.conversation_id = conversations.id").
		Where("p.user_id = ?", userID).
		Find(&obj).Error
	if err != nil {
		return nil, err
	}

	return obj, nil
}
