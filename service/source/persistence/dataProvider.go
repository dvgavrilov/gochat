package persistence

import (
	"io"

	"github.com/dvgavrilov/gochat/service/source/interfaces"
	"github.com/dvgavrilov/gochat/service/source/persistence/database"
)

// var dataSource interfaces.DataSource
var dsNew *database.DataSourceNew

// Init func
func Init() error {

	if dsNew != nil {
		return nil
	}

	dsNew = &database.DataSourceNew{}
	err := dsNew.Init()

	return err
}

// DataSourceCloser func
func DataSourceCloser() io.Closer {
	Init()
	return dsNew
}

// GetConversationsProvider func
func GetConversationsProvider() interfaces.ConversationsProvider {
	Init()
	return dsNew.ConversationsManager
}

// GetMessagesProvider func
func GetMessagesProvider() interfaces.MessagesProvider {
	Init()
	return dsNew.MessagesManager
}

// GetParticipantsProvider func
func GetParticipantsProvider() interfaces.ParticipantsProvider {
	Init()
	return dsNew.ParticipantStore
}

// GetUnreadInfoManager func
func GetUnreadInfoManager() interfaces.UnreadInfoManager {
	Init()
	return dsNew.UnreadInfoStore
}
