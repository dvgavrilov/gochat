package database

import "leto-yanao-1/service/source/models"

type unreadInfoDataStore struct {
	connection *Connection
}

func (r unreadInfoDataStore) Add(unread *models.UnreadInfo) error {
	err := r.connection.db.Create(unread).Error
	if err != nil {
		return err
	}

	return nil
}

func (r unreadInfoDataStore) MarkAsRead(messageID uint, participantID uint) error {
	return r.connection.db.Model(&models.UnreadInfo{}).Where("message_id = ? AND participant_id = ?", messageID, participantID).Update("read", true).Error
}

// GetByMessageID func
func (r unreadInfoDataStore) GetByMessageID(messageID uint) (*[]models.UnreadInfo, error) {
	obj := &[]models.UnreadInfo{}
	err := r.connection.db.Where("message_id = ?", messageID).Find(obj).Error
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (r unreadInfoDataStore) GetForUser(participantID uint) (int, error) {
	var count int
	err := r.connection.db.Model(&models.UnreadInfo{}).
		Select("count(distinct(message_id))").
		Where(
			`
			read = false 
			AND
			participant_id = ?
			`, participantID).Count(&count).Error

	return count, err
}

func (r unreadInfoDataStore) GetForUserAndGlobal(participantID uint) (int, error) {
	var count int
	err := r.connection.db.Model(&models.UnreadInfo{}).
		Select("count(distinct(message_id))").
		Where(
			`
			read = false 
			AND
			(participant_id = 0 OR participant_id = ?)
			`, participantID).Count(&count).Error

	return count, err
}
