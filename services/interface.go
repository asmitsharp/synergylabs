package services

import (
	"context"
	"mime/multipart"
	"synergylabs/models"
)

type UserServiceInterface interface {
	CreateUser(ctx context.Context, user *models.User) error
	ValidateLogin(ctx context.Context, email, password string) (*models.User, error)
	GetAllApplicants(ctx context.Context, page, pageSize int) (*PaginatedResponse, error)
	GetApplicantWithProfile(ctx context.Context, id uint) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uint) error
}

type JobServiceInterface interface {
	CreateJob(ctx context.Context, job *models.Job) error
	GetJobs(ctx context.Context, filters JobFilters) (*PaginatedResponse, error)
	GetJobWithApplicants(ctx context.Context, id uint) (*models.Job, error)
	ApplyToJob(ctx context.Context, jobID, userID uint) error
	UpdateJob(ctx context.Context, job *models.Job) error
	DeleteJob(ctx context.Context, id uint) error
}

type ResumeServiceInterface interface {
	ProcessResume(ctx context.Context, file *multipart.FileHeader, userID uint) error
	GetResumeData(ctx context.Context, userID uint) (*models.Profile, error)
}
