package services

import (
	"context"
	"errors"
	"fmt"
	"synergylabs/models"
	"synergylabs/services/cache"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db     *gorm.DB
	cache  *cache.Cache
	logger *zap.Logger
}

var _ UserServiceInterface = (*UserService)(nil)

func NewUserService(db *gorm.DB, cache *cache.Cache, logger *zap.Logger) *UserService {
	return &UserService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(tx.Error))
		return tx.Error
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		s.logger.Error("Failed to hash password", zap.Error(err))
		return err
	}
	user.PasswordHash = string(hashedPassword)

	// Create user
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create user", zap.Error(err))
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.New("email already exists")
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		return err
	}

	// Invalidate cache
	s.cache.Delete(ctx, string(ApplicantsCacheKey))

	s.logger.Info("User created successfully", zap.Uint("user_id", user.ID))

	return nil
}

func (s *UserService) GetAllApplicants(ctx context.Context, page, pageSize int) (*PaginatedResponse, error) {

	cacheKey := fmt.Sprintf("%s:%d:%d", ApplicantsCacheKey, page, pageSize)
	var response PaginatedResponse

	// Try to get from cache
	err := s.cache.Get(ctx, cacheKey, &response)
	if err == nil {
		return &response, nil
	}

	// Get total count
	var total int64
	if err := s.db.Model(&models.User{}).
		Where("user_type = ?", models.UserTypeApplicant).
		Count(&total).Error; err != nil {
		s.logger.Error("Failed to count applicants", zap.Error(err))
		return nil, err
	}

	// Get paginated applicants
	var applicants []models.User
	if err := s.db.WithContext(ctx).
		Where("user_type = ?", models.UserTypeApplicant).
		Preload("Profile"). // Eager load profiles
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&applicants).Error; err != nil {
		s.logger.Error("Failed to fetch applicants", zap.Error(err))
		return nil, err
	}

	response = PaginatedResponse{
		Data:       applicants,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(total)/pageSize + 1,
	}

	// Cache the response
	if err := s.cache.Set(ctx, cacheKey, response, 5*time.Minute); err != nil {
		s.logger.Warn("Failed to cache applicants", zap.Error(err))
	}

	return &response, nil
}

func (s *UserService) ValidateLogin(ctx context.Context, email, password string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		s.logger.Error("User not found", zap.Error(err))
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Error("Invalid password", zap.Error(err))
		return nil, errors.New("invalid password")
	}

	return &user, nil
}

func (s *UserService) GetApplicantWithProfile(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Preload("Profile").First(&user, id).Error; err != nil {
		s.logger.Error("Failed to fetch applicant with profile", zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	if err := s.db.WithContext(ctx).Save(user).Error; err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return err
	}
	s.logger.Info("User updated successfully", zap.Uint("user_id", user.ID))
	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Delete(&models.User{}, id).Error; err != nil {
		s.logger.Error("Failed to delete user", zap.Error(err))
		return err
	}
	s.logger.Info("User deleted successfully", zap.Uint("user_id", id))
	return nil
}
