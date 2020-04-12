package models

import "time"

const (
	StatusNew  = 0
	StatusRead = 1
	//StatusDeleted = 2 not required by today

	ContentText  = 1 // text
	ContentImage = 2 // image
)

// Message struct
type Message struct {
	ID             uint
	SenderID       uint
	ConversationID uint
	ApplicationID  uint
	ContentType    uint
	UnreadInfo     []UnreadInfo
	Read           bool `gorm:"-"`
	Content        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// UnreadInfo struct
type UnreadInfo struct {
	MessageID      uint
	ConversationID uint
	ParticipantID  uint
	Read           bool
}

// Participant struct
type Participant struct {
	UserID         uint
	ConversationID uint
}

// Conversation struct
type Conversation struct {
	ID            uint
	Participants  []Participant
	ApplicationID uint
	CreatedAt     time.Time
	UpdatedAt     time.Time
	UnreadCount   uint `gorm:"-"`
}
