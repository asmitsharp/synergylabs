package services

import (
	"context"
	"errors"
	"fmt"
	"synergylabs/models"
	"synergylabs/services/cache"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type JobService struct {
	db     *gorm.DB
	cache  *cache.Cache
	logger *zap.Logger
}

func NewJobService(db *gorm.DB, cache *cache.Cache, logger *zap.Logger) *JobService {
	return &JobService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

func (s *JobService) GetJobs(ctx context.Context, filters JobFilters) (*PaginatedResponse, error) {

	cacheKey := fmt.Sprintf("%s:%s:%s:%d:%d",
		JobsCacheKey,
		filters.Title,
		filters.CompanyName,
		filters.Page,
		filters.PageSize,
	)

	var response PaginatedResponse

	// Try to get from cache
	err := s.cache.Get(ctx, cacheKey, &response)
	if err == nil {
		return &response, nil
	}

	// Build query
	query := s.db.WithContext(ctx).Model(&models.Job{})

	if filters.Title != "" {
		query = query.Where("title ILIKE ?", "%"+filters.Title+"%")
	}
	if filters.CompanyName != "" {
		query = query.Where("company_name ILIKE ?", "%"+filters.CompanyName+"%")
	}
	if !filters.PostedAfter.IsZero() {
		query = query.Where("posted_on >= ?", filters.PostedAfter)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count jobs", zap.Error(err))
		return nil, err
	}

	// Get paginated jobs
	var jobs []models.Job
	if err := query.
		Preload("PostedBy"). // Eager load related data
		Offset((filters.Page - 1) * filters.PageSize).
		Limit(filters.PageSize).
		Find(&jobs).Error; err != nil {
		s.logger.Error("Failed to fetch jobs", zap.Error(err))
		return nil, err
	}

	totalPages := 1
	if filters.PageSize > 0 {
		totalPages = int(total)/filters.PageSize + 1
	}

	response = PaginatedResponse{
		Data:       jobs,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}

	// Cache the response
	if err := s.cache.Set(ctx, cacheKey, response, 5*time.Minute); err != nil {
		s.logger.Warn("Failed to cache jobs", zap.Error(err))
	}

	return &response, nil
}

func (s *JobService) ApplyToJob(ctx context.Context, jobID, userID uint) error {

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(tx.Error))
		return tx.Error
	}

	// Check if already applied
	var count int64
	if err := tx.Model(&models.Job{}).
		Where("id = ?", jobID).
		Joins("JOIN job_applications ON job_applications.job_id = jobs.id").
		Where("job_applications.user_id = ?", userID).
		Count(&count).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to check existing application", zap.Error(err))
		return err
	}

	if count > 0 {
		tx.Rollback()
		return errors.New("already applied to this job")
	}

	// Apply to job
	if err := tx.Model(&models.Job{}).
		Where("id = ?", jobID).
		Association("Applicants").
		Append(&models.User{Model: gorm.Model{ID: userID}}); err != nil {
		tx.Rollback()
		s.logger.Error("Failed to apply to job", zap.Error(err))
		return err
	}

	// Update total applications count
	if err := tx.Model(&models.Job{}).
		Where("id = ?", jobID).
		UpdateColumn("total_applications", gorm.Expr("total_applications + ?", 1)).
		Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update applications count", zap.Error(err))
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		return err
	}

	// Invalidate cache
	s.cache.Delete(ctx, string(JobsCacheKey))

	s.logger.Info("Successfully applied to job",
		zap.Uint("job_id", jobID),
		zap.Uint("user_id", userID),
	)

	return nil
}

func (s *JobService) CreateJob(ctx context.Context, job *models.Job) error {
	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		s.logger.Error("Failed to create job", zap.Error(err))
		return err
	}
	s.logger.Info("Job created successfully", zap.Uint("job_id", job.ID))

	// Invalidate cache for jobs
	s.cache.Delete(ctx, string(JobsCacheKey))

	return nil
}

func (s *JobService) GetJobWithApplicants(ctx context.Context, id uint) (*models.Job, error) {
	cacheKey := fmt.Sprintf("%s:%d", JobsCacheKey, id)
	var job models.Job

	// Try to get from cache
	err := s.cache.Get(ctx, cacheKey, &job)
	if err == nil {
		return &job, nil
	}

	// If not found in cache, fetch from database
	if err := s.db.WithContext(ctx).Preload("Applicants").First(&job, id).Error; err != nil {
		s.logger.Error("Failed to fetch job with applicants", zap.Error(err))
		return nil, err
	}

	// Cache the job with applicants
	if err := s.cache.Set(ctx, cacheKey, job, 5*time.Minute); err != nil {
		s.logger.Warn("Failed to cache job with applicants", zap.Error(err))
	}

	return &job, nil
}

func (s *JobService) UpdateJob(ctx context.Context, job *models.Job) error {
	if err := s.db.WithContext(ctx).Save(job).Error; err != nil {
		s.logger.Error("Failed to update job", zap.Error(err))
		return err
	}
	s.logger.Info("Job updated successfully", zap.Uint("job_id", job.ID))

	// Invalidate cache for the specific job
	cacheKey := fmt.Sprintf("%s:%d", JobsCacheKey, job.ID)
	s.cache.Delete(ctx, cacheKey)

	return nil
}

func (s *JobService) DeleteJob(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Delete(&models.Job{}, id).Error; err != nil {
		s.logger.Error("Failed to delete job", zap.Error(err))
		return err
	}
	s.logger.Info("Job deleted successfully", zap.Uint("job_id", id))

	// Invalidate cache for the deleted job
	cacheKey := fmt.Sprintf("%s:%d", JobsCacheKey, id)
	s.cache.Delete(ctx, cacheKey)

	return nil
}
