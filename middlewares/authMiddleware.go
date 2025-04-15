package middlewares

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		role := session.Get("role")
		c.Set("role", role)
		c.Next()
	}
}

func ApplicantOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role, exists := c.Get("role"); !exists || role.(string) != "applicant" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Applicants only"})
			return
		}
		c.Next()
	}
}

func RecruiterOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role, exists := c.Get("role"); !exists || role.(string) != "recruiter" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Recruiters only"})
			return
		}
		c.Next()
	}
}

func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role, exists := c.Get("role"); !exists || role.(string) != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admins only"})
			return
		}
		c.Next()
	}
}
