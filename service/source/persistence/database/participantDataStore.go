package database

import (
	"leto-yanao-1/service/source/models"
)

type participantDataStore struct {
	connection *Connection
}

func (r participantDataStore) Add(participant *models.Participant) error {
	err := r.connection.db.Create(&participant).Error
	if err != nil {
		return err
	}
	return nil
}

func (r participantDataStore) GetByConversationID(conversationID uint) (*[]models.Participant, error) {
	obj := &[]models.Participant{}
	err := r.connection.db.Where("conversation_id = ?", conversationID).Find(obj).Error
	if err != nil {
		return nil, err
	}

	return obj, nil
}
