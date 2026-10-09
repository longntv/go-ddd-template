package datastore

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/longntv/go-ddd-template/internal/domain/entity"
)

// UserEntity represents the users table.
type UserEntity struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `gorm:"type:varchar(255);not null;default:''"`
	Email     string    `gorm:"type:varchar(255);not null;default:'';uniqueIndex"`
	Password  string    `gorm:"type:varchar(255);not null;default:''"`
	CreatedAt time.Time `gorm:"type:timestamp(3);not null;default:CURRENT_TIMESTAMP(3)"`
	UpdatedAt time.Time `gorm:"type:timestamp(3);not null;default:CURRENT_TIMESTAMP(3)"`
}

// TableName specifies the table name for UserEntity.
func (UserEntity) TableName() string {
	return "users"
}

// BeforeCreate hook.
func (e *UserEntity) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// ToDomain converts UserEntity to domain entity.User.
func (e *UserEntity) ToDomain() *entity.User {
	return &entity.User{
		ID:        entity.UserID(e.ID),
		Name:      e.Name,
		Email:     e.Email,
		Password:  e.Password,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}
