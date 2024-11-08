package api

import (
	"net/http"
	"strconv"
	"synergylabs/models"
	"synergylabs/services"
	"synergylabs/services/cache"
	"synergylabs/util"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	userService   services.UserService
	jobService    services.JobService
	resumeService services.ResumeService
)

// SetupRoutes initializes the API routes
func SetupRoutes(e *echo.Echo, db *gorm.DB, redisCache *cache.Cache, logger *zap.Logger) {
	userService = *services.NewUserService(db, redisCache, logger)
	jobService = *services.NewJobService(db, redisCache, logger)
	resumeService = *services.NewResumeService(db, logger)

	// User routes
	e.POST("/signup", Signup)
	e.POST("/login", Login)

	// Resume routes
	e.POST("/uploadResume", UploadResume, util.AuthMiddleware, util.ApplicantOnly)

	// Job routes
	e.POST("/admin/job", CreateJob, util.AuthMiddleware, util.AdminOnly)
	e.GET("/admin/job/:job_id", GetJobWithApplicants, util.AuthMiddleware, util.AdminOnly)
	e.GET("/admin/applicants", GetAllApplicants, util.AuthMiddleware, util.AdminOnly)
	e.GET("/admin/applicant/:applicant_id", GetApplicantData, util.AuthMiddleware, util.AdminOnly)

	// Public job routes
	e.GET("/jobs", GetJobs, util.AuthMiddleware)
	e.GET("/jobs/apply", ApplyToJob, util.AuthMiddleware, util.ApplicantOnly)
}

// Signup handles user registration
func Signup(c echo.Context) error {
	var user models.User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	// Validate user input
	if user.Name == "" || user.Email == "" || user.PasswordHash == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name, email, and password are required"})
	}

	if err := userService.CreateUser(c.Request().Context(), &user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "User created successfully"})
}

// Login handles user authentication
func Login(c echo.Context) error {
	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.Bind(&loginData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	// Validate login data
	if loginData.Email == "" || loginData.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email and password are required"})
	}

	user, err := userService.ValidateLogin(c.Request().Context(), loginData.Email, loginData.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	// Generate JWT token
	token, err := util.GenerateToken(user.ID, string(user.UserType))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not generate token"})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

// UploadResume handles resume upload
func UploadResume(c echo.Context) error {
	userID := c.Get("userId").(uint)
	file, err := c.FormFile("resume")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid file"})
	}

	if err := resumeService.ProcessResume(c.Request().Context(), file, userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Resume uploaded successfully"})
}

// CreateJob handles job creation
func CreateJob(c echo.Context) error {
	var job models.Job
	if err := c.Bind(&job); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	// Validate job data
	if job.Title == "" || job.Description == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Job title and description are required"})
	}

	if err := jobService.CreateJob(c.Request().Context(), &job); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Job created successfully"})
}

// GetJobWithApplicants retrieves job details and applicants
func GetJobWithApplicants(c echo.Context) error {
	jobID := c.Param("job_id")
	id, err := strconv.ParseUint(jobID, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid job ID")
	}
	job, err := jobService.GetJobWithApplicants(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, job)
}

// GetAllApplicants retrieves all applicants
func GetAllApplicants(c echo.Context) error {
	applicants, err := userService.GetAllApplicants(c.Request().Context(), 1, 100) // Example pagination
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, applicants)
}

// GetApplicantData retrieves specific applicant data
func GetApplicantData(c echo.Context) error {
	applicantID := c.Param("applicant_id")
	id, err := strconv.ParseUint(applicantID, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid Applicant ID")
	}
	applicant, err := userService.GetApplicantWithProfile(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, applicant)
}

// GetJobs retrieves job openings
func GetJobs(c echo.Context) error {
	jobs, err := jobService.GetJobs(c.Request().Context(), services.JobFilters{}) // Add filters as needed
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, jobs)
}

// ApplyToJob handles job applications
func ApplyToJob(c echo.Context) error {
	userID := c.Get("userId").(uint)
	jobID := c.QueryParam("job_id")

	if jobID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Job ID is required"})
	}

	id, err := strconv.ParseUint(jobID, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid job ID")
	}

	if err := jobService.ApplyToJob(c.Request().Context(), uint(id), userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Applied to job successfully"})
}
