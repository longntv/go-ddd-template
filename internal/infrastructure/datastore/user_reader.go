package datastore

import (
	"context"

	"gorm.io/gorm"

	"go-ddd-template/internal/domain/entity"
	"go-ddd-template/internal/domain/gateway"
	"go-ddd-template/internal/domain/model"
)

// userReader implements gateway.UserQueriesGateway interface.
type userReader struct {
	db *gorm.DB
}

// NewUserReader creates a new Reader.
func NewUserReader(db *gorm.DB) gateway.UserQueriesGateway {
	return &userReader{db: db}
}

// Get retrieves a user by ID.
func (r *userReader) Get(ctx context.Context, id entity.UserID) (*entity.User, error) {
	var userEntity UserEntity
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&userEntity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	return userEntity.ToDomain(), nil
}

// GetByEmail retrieves a user by email.
func (r *userReader) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var userEntity UserEntity
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&userEntity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	return userEntity.ToDomain(), nil
}

// List retrieves a list of users with pagination.
func (r *userReader) List(ctx context.Context, limit, offset int) ([]*entity.User, int, error) {
	var userEntities []*UserEntity
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).Model(&UserEntity{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&userEntities).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*entity.User, 0, len(userEntities))
	for _, e := range userEntities {
		users = append(users, e.ToDomain())
	}

	return users, int(total), nil
}

// Exists checks if a user exists by email.
func (r *userReader) Exists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserEntity{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
