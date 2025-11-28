package modeluser

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:255;not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;size:255;not null"`
	Password  string    `json:"-" gorm:"column:password_hash;size:255;not null"`
	Role      string    `json:"role" gorm:"type:user_role;default:'staff'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
