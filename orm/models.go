package orm

import (
	"time"
)

type Record struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	Imei      string    `gorm:"not null;index" json:"imei"`
	Payload   string    `gorm:"not null" json:"payload"`
}
