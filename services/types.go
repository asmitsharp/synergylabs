package services

import (
	"time"
)

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

type JobFilters struct {
	Title       string    `json:"title"`
	CompanyName string    `json:"company_name"`
	PostedAfter time.Time `json:"posted_after"`
	Page        int       `json:"page"`
	PageSize    int       `json:"page_size"`
}

type CacheKey string

const (
	JobsCacheKey       CacheKey = "jobs"
	ApplicantsCacheKey CacheKey = "applicants"
)
