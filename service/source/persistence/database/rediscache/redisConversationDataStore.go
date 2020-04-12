package rediscache

import (
	"encoding/json"
	"fmt"

	"log"

	"github.com/dvgavrilov/gochat/service/source/config"
	"github.com/dvgavrilov/gochat/service/source/customerrors"
	"github.com/dvgavrilov/gochat/service/source/interfaces"
	"github.com/dvgavrilov/gochat/service/source/models"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
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
	chatSeqPrefix           = "seq"
)

var pool *redis.Pool

type redisConversationDataStore struct {
	conversationsManager interfaces.ConversationsProvider
}

// GetRedisConversationStore func
func GetRedisConversationStore(store interfaces.ConversationsProvider) (interfaces.ConversationsProvider, error) {
	if store == nil {
		return nil, customerrors.ErrArgumentInvalid
	}

	pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", fmt.Sprintf(
				"%s:%v",
				config.MainConfiguration.RedisSettings.Address,
				config.MainConfiguration.RedisSettings.Port))
			if err != nil {
				log.Panic(err)
			}

			return conn, nil
		}}

	r := &redisConversationDataStore{
		conversationsManager: store,
	}

	return r, nil
}

// Close func
func Close() error {
	if pool == nil {
		return nil
	}

	return pool.Close()
}

func (r redisConversationDataStore) GetByUserID(userID uint) (*[]models.Conversation, error) {
	return r.conversationsManager.GetByUserID(userID)
}
func (r redisConversationDataStore) GetByApplicationID(applicationID uint) (*models.Conversation, error) {
	conn := pool.Get()
	defer conn.Close()

	key := chatApplicationToConversationKey(applicationID)
	cid, err := redis.Int(conn.Do("GET", key))
	if err != nil {
		logrus.Error(err)

		conv, err := r.conversationsManager.GetByApplicationID(applicationID)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}

		redisSetConversation(conn, conv)

		return conv, nil
	}

	ret := &models.Conversation{}
	err = redisGet(conn, chatConversationID(uint(cid)), ret)
	if err != nil {
		logrus.Error(err)

		conv, err := r.conversationsManager.GetByApplicationID(applicationID)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}

		redisSetConversation(conn, conv)

		return conv, nil
	}

	return ret, nil
}

func (r redisConversationDataStore) Add(conversation *models.Conversation) (*models.Conversation, error) {

	conn := pool.Get()
	defer conn.Close()

	ret, err := r.conversationsManager.Add(conversation)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	err = redisSetConversation(conn, ret)
	if err != nil {
		logrus.Error(err)
	}

	return r.conversationsManager.Add(conversation)
}

func redisGet(conn redis.Conn, key string, dest interface{}) error {

	objStr, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return err
	}
	b := []byte(objStr)

	err = json.Unmarshal(b, dest)
	return err
}

func getListOfIntegers(conn redis.Conn, key string) (*[]int, error) {

	values, err := redis.Values(conn.Do("LRANGE", key, 0, -1))
	if err != nil {
		return nil, err
	}

	ret := make([]int, 0)
	err = redis.ScanSlice(values, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func redisSetConversation(conn redis.Conn, conv *models.Conversation) error {
	raw, err := json.Marshal(conv)
	if err != nil {
		return err
	}

	key := chatConversationID(conv.ID)
	_, err = conn.Do("SET", key, string(raw))
	if err != nil {
		logrus.Error(err)
		return err
	}

	key = chatApplicationToConversationKey(conv.ApplicationID)
	_, err = conn.Do("SET", key, conv.ID)
	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

// chat.conversation.{id}
func chatConversationID(conversationID uint) string {
	return fmt.Sprintf("%s.%s.%v", chatPrefix, conversationPrefix, conversationID)
}

// chat.application.conversation.{id}
func chatApplicationToConversationKey(applicationID uint) string {
	return fmt.Sprintf("%s.%s.%s.%v", chatPrefix, chatApplicationPrefix, conversationPrefix, applicationID)
}
