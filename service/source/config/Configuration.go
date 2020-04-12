package config

import (
	"encoding/json"
	"os"
)

// Configuration стуктура
type Configuration struct {
	DatabaseSettings    DatabaseSettings
	WebSocketSettings   WebSocketSettings
	ChatRoomSettings    ChatRoomSettings
	RedisSettings       RedisSettings
	ApplicationSettings ApplicationSettings
}

// DatabaseSettings стуктура
type DatabaseSettings struct {
	Engine   string
	Server   string
	Port     string
	User     string
	Password string
	Database string
}

// WebSocketSettings struct
type WebSocketSettings struct {
	ReadBufferSize  int
	WriteBufferSize int
	Port            int
	Origin          string
}

// ChatRoomSettings struct
type ChatRoomSettings struct {
	MaxValueOfChatters int
	CacheHistory       string
}

// ApplicationSettings struct
type ApplicationSettings struct {
	DebugMode string
}

// RedisSettings struct
type RedisSettings struct {
	Address string
	Port    int
}

// Read функция отвечает за загрузка конфигурации с json файла и декодирования его структуру Configuration
func Read() error {

	file, err := os.Open("../config/configuration.json")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&MainConfiguration)
	if err != nil {
		return err
	}

	return nil
}

var MainConfiguration Configuration
