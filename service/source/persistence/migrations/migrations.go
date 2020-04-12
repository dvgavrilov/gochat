package migrations

import (
	"leto-yanao-1/service/source/models"

	"github.com/jinzhu/gorm"
)

func Migrate(db *gorm.DB) error {

	db.AutoMigrate(&models.Message{}, &models.Conversation{}, &models.Participant{}, &models.UnreadInfo{})

	db.Model(&models.Conversation{}).AddUniqueIndex("idx_id_appid", "id", "application_id")
	db.Model(&models.Conversation{}).AddUniqueIndex("idx_appid", "application_id")

	db.Model(&models.Message{}).AddForeignKey("conversation_id", "conversations(id)", "RESTRICT", "RESTRICT")
	db.Model(&models.Message{}).AddForeignKey("application_id", "conversations(application_id)", "RESTRICT", "RESTRICT")

	db.Model(&models.Participant{}).AddForeignKey("conversation_id", "conversations(id)", "RESTRICT", "RESTRICT")

	db.Model(&models.UnreadInfo{}).AddForeignKey("conversation_id", "conversations(id)", "RESTRICT", "RESTRICT")
	db.Model(&models.UnreadInfo{}).AddForeignKey("message_id", "messages(id)", "RESTRICT", "RESTRICT")

	db.Model(&models.Participant{}).AddUniqueIndex("idx_convid_userid", "conversation_id", "user_id")

	return nil
}
