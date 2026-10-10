package datastore

import (
	"context"

	"gorm.io/gorm"

	"github.com/longntv/go-ddd-template/internal/domain/entity"
	"github.com/longntv/go-ddd-template/internal/domain/gateway"
)

// userWriter implements gateway.UserCommandsGateway interface.
type userWriter struct {
	db *gorm.DB
}

// NewUserWriter creates a new Writer.
func NewUserWriter(db *gorm.DB) gateway.UserCommandsGateway {
	return &userWriter{db: db}
}

// Create creates a new user.
func (w *userWriter) Create(ctx context.Context, user *entity.User) error {
	userEntity := &UserEntity{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}
	return conn(ctx, w.db).Create(userEntity).Error
}

// Update updates an existing user.
func (w *userWriter) Update(ctx context.Context, user *entity.User) error {
	updates := map[string]interface{}{
		"name":          user.Name,
		"email":         user.Email,
		"password_hash": user.PasswordHash,
	}
	result := conn(ctx, w.db).Model(&UserEntity{}).Where("id = ?", user.ID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete deletes a user by ID.
func (w *userWriter) Delete(ctx context.Context, id entity.UserID) error {
	result := conn(ctx, w.db).Delete(&UserEntity{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
