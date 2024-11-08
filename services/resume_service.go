package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"synergylabs/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ResumeService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewResumeService(db *gorm.DB, logger *zap.Logger) *ResumeService {
	return &ResumeService{
		db:     db,
		logger: logger,
	}
}

func (s *ResumeService) ProcessResume(ctx context.Context, file *multipart.FileHeader, userID uint) error {
	// Open the file
	f, err := file.Open()
	if err != nil {
		s.logger.Error("Failed to open resume file", zap.Error(err))
		return err
	}
	defer f.Close()

	// Read the file into a byte buffer
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(f); err != nil {
		s.logger.Error("Failed to read resume file", zap.Error(err))
		return err
	}

	// Send the file to the third-party API
	apiURL := "https://api.apilayer.com/resume_parser/upload"
	req, err := http.NewRequest("POST", apiURL, buf)
	if err != nil {
		s.logger.Error("Failed to create request", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("apikey", "0bWeisRWoLj3UdXt3MXMSMWptYFIpQfS") // Use a secure way to manage API keys

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Failed to send request to resume parser API", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Failed to parse resume, received non-200 response", zap.Int("status_code", resp.StatusCode))
		return fmt.Errorf("failed to parse resume, status code: %d", resp.StatusCode)
	}

	// Parse the response
	var resumeData struct {
		Name      string   `json:"name"`
		Email     string   `json:"email"`
		Phone     string   `json:"phone"`
		Skills    []string `json:"skills"`
		Education []struct {
			Name string `json:"name"`
		} `json:"education"`
		Experience []struct {
			Name string `json:"name"`
		} `json:"experience"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&resumeData); err != nil {
		s.logger.Error("Failed to decode resume data", zap.Error(err))
		return err
	}

	// Prepare the profile data
	profile := models.Profile{
		ApplicantID: userID,
		Name:        resumeData.Name,
		Email:       resumeData.Email,
		Phone:       resumeData.Phone,
		Skills:      fmt.Sprintf("%v", resumeData.Skills),     // Convert slice to string
		Education:   fmt.Sprintf("%v", resumeData.Education),  // Convert slice to string
		Experience:  fmt.Sprintf("%v", resumeData.Experience), // Convert slice to string
	}

	// Save the profile to the database
	if err := s.db.WithContext(ctx).Save(&profile).Error; err != nil {
		s.logger.Error("Failed to save profile data", zap.Error(err))
		return err
	}

	s.logger.Info("Resume processed and profile updated successfully", zap.Uint("user_id", userID))
	return nil
}

func (s *ResumeService) GetResumeData(ctx context.Context, userID uint) (*models.Profile, error) {
	var profile models.Profile
	if err := s.db.WithContext(ctx).Where("applicant_id = ?", userID).First(&profile).Error; err != nil {
		s.logger.Error("Failed to fetch resume data", zap.Error(err))
		return nil, err
	}
	return &profile, nil
}
