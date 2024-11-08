package util

import (
	"net/http"
	"strings"
	"synergylabs/models"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No token provided"})
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
		claims, err := ValidateToken(token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}

		c.Set("userId", claims.UserId)
		c.Set("userType", claims.UserType)
		return next(c)
	}
}

func AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userType := c.Get("userType").(string)
		if userType != string(models.UserTypeAdmin) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin access required"})
		}
		return next(c)
	}
}

func ApplicantOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userType := c.Get("userType").(string)
		if userType != string(models.UserTypeApplicant) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Applicant access required"})
		}
		return next(c)
	}
}
