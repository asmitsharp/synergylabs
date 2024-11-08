package models

import (
	"time"

	"gorm.io/gorm"
)

type UserType string

const (
	UserTypeAdmin     UserType = "ADMIN"
	UserTypeApplicant UserType = "APPLICANT"
)

type User struct {
	gorm.Model
	Name            string   `json:"name"`
	Email           string   `json:"email" gorm:"unique"`
	Address         string   `json:"address"`
	UserType        UserType `json:"user_type"`
	PasswordHash    string   `json:"password_hash"`
	ProfileHeadline string   `json:"profile_headline"`
	Profile         *Profile `json:"profile,omitempty" gorm:"foreignKey:ApplicantID"`
}

type Profile struct {
	gorm.Model
	ApplicantID       uint   `json:"applicant_id"`
	ResumeFileAddress string `json:"resume_file_address"`
	Skills            string `json:"skills"`
	Education         string `json:"education"`
	Experience        string `json:"experience"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
}

type Job struct {
	gorm.Model
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	PostedOn          time.Time `json:"posted_on"`
	TotalApplications int       `json:"total_applications"`
	CompanyName       string    `json:"company_name"`
	PostedByID        uint      `json:"posted_by_id"`
	PostedBy          User      `json:"posted_by" gorm:"foreignKey:PostedByID"`
	Applicants        []User    `json:"applicants,omitempty" gorm:"many2many:job_applications;"`
}
