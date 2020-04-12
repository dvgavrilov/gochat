package database

import (
	"leto-yanao-1/service/source/models"
)

type messagesDataStore struct {
	connection *Connection
}

type messagesManager struct {
	messagesStore   messagesDataStore
	unreadInfoStore unreadInfoDataStore
}

func (r *messagesManager) Init(connection *Connection) {
	r.messagesStore = messagesDataStore{connection: connection}
	r.unreadInfoStore = unreadInfoDataStore{connection: connection}
}

func (r messagesManager) GetMessages(conversationID uint) (*[]models.Message, error) {
	ret, err := r.messagesStore.GetMessages(conversationID)
	if err != nil {
		return nil, err
	}

	return r.populateUnreadInfo(ret)
}

func (r messagesManager) GetUnreadMessages(userID uint) (*[]models.Message, error) {
	ret, err := r.messagesStore.GetUnreadMessages(userID)
	if err != nil {
		return nil, err
	}

	return r.populateUnreadInfo(ret)
}

func (r messagesManager) AddMessage(message *models.Message) (*models.Message, error) {
	return r.messagesStore.AddMessage(message)
}

// GetMessages func
func (r messagesDataStore) GetMessages(conversationID uint) (*[]models.Message, error) {
	obj := &[]models.Message{}
	err := r.connection.db.Where("conversation_id = ?", conversationID).Find(obj).Error
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (r messagesDataStore) GetUnreadMessages(userID uint) (*[]models.Message, error) {
	obj := &[]models.Message{}
	err := r.connection.db.Model(models.Message{}).
		Select("distinct on(messages.id) messages.id, sender_id, messages.conversation_id, content_type, content, messages.created_at, messages.updated_at, messages.application_id").
		Joins("join conversations c on c.id= messages.conversation_id ").
		Joins("join unread_infos ui on ui.message_id = messages.id").
		Where("ui.participant_id IN (0, ?) AND ui.read = false", userID).
		Find(obj).Error

	if err != nil {
		return nil, err
	}

	return obj, nil
}

// AddMessage func
func (r messagesDataStore) AddMessage(message *models.Message) (*models.Message, error) {
	r.connection.db.NewRecord(message)
	err := r.connection.db.Create(&message).Error
	if err != nil {
		return nil, err
	}

	return message, err
}

func (r messagesManager) populateUnreadInfo(messages *[]models.Message) (*[]models.Message, error) {

	for i := range *messages {
		infos, err := r.unreadInfoStore.GetByMessageID((*messages)[i].ID)
		if err != nil {
			return nil, err
		}

		(*messages)[i].UnreadInfo = *infos
	}

	return messages, nil
}
