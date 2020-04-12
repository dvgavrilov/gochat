package database

import (
	"errors"
	"fmt"
	"io"
	"leto-yanao-1/service/source/config"
	"leto-yanao-1/service/source/interfaces"
	"leto-yanao-1/service/source/persistence/database/rediscache"
	"leto-yanao-1/service/source/persistence/migrations"
	"strconv"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

const (
	chanListKey             = "channel.list"
	historyPrefix           = "history"
	conversationPrefix      = "conversation"
	customerPrefix          = "customer"
	moderatorPrefix         = "moderator"
	messagePrefix           = "message"
	chatPrefix              = "chat"
	chatLastSeqPrefix       = "last_seq"
	chatApplicationPrefix   = "application"
	chatClientLastSeqPrefix = "client.last_seq"

	chatSeqPrefix = "seq"

	maxHistorySize int64 = 1000
)

var (
	dbConnection *gorm.DB
	useCache     bool
)

// Connection struct
type Connection struct {
	db *gorm.DB
}

// DataSourceNew struct
type DataSourceNew struct {
	connection           *Connection
	MessagesManager      interfaces.MessagesProvider
	ConversationsManager interfaces.ConversationsProvider
	ParticipantStore     interfaces.ParticipantsProvider
	UnreadInfoStore      interfaces.UnreadInfoManager

	io.Closer
}

// Close func
func (r *DataSourceNew) Close() error {
	if useCache {
		rediscache.Close()
	}

	return r.connection.db.Close()
}

// Init func
func (r *DataSourceNew) Init() error {

	var err error
	dbConnection, err = gorm.Open("postgres", getConnString())
	if err != nil {
		return err
	}

	(*r).connection = &Connection{dbConnection}

	migrations.Migrate(dbConnection)

	useCache, err = strconv.ParseBool(config.MainConfiguration.ChatRoomSettings.CacheHistory)
	if err != nil {
		return errors.New("error with reading config file")
	}

	msgManager := &messagesManager{}
	msgManager.Init(r.connection)

	r.MessagesManager = msgManager
	r.ParticipantStore = participantDataStore{connection: r.connection}

	store := conversationsManager{}
	store.Init(r.connection)

	r.UnreadInfoStore = unreadInfoDataStore{connection: r.connection}

	if useCache {
		redisStore, err := rediscache.GetRedisConversationStore(store)
		if err != nil {
			return nil
		}

		r.ConversationsManager = redisStore

	} else {
		r.ConversationsManager = store
	}

	return nil
}

func getConnString() string {

	return fmt.Sprintf("host=%v user=%v dbname=%v sslmode=disable password=%v",
		config.MainConfiguration.DatabaseSettings.Server,
		config.MainConfiguration.DatabaseSettings.User,
		config.MainConfiguration.DatabaseSettings.Database,
		config.MainConfiguration.DatabaseSettings.Password)
}
