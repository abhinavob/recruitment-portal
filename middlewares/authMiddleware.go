package middlewares

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		// token := session.Get("token")

		// if token == nil {
		// 	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		// 	return
		// }

		// // Verify the session in the database
		// dbSession, err := queries.GetSession(context.Background(), token.(string))
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		// 	return
		// }

		// // Check if the session is too old (e.g., older than 24 hours)
		// if time.Since(dbSession.CreatedAt.Time) > 24*time.Hour {
		// 	// Delete the expired session
		// 	queries.DeleteSession(context.Background(), db.DeleteSessionParams{
		// 		Token: token.(string),
		// 	})
		// 	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
		// 	return
		// }

		// // Set the role in the context for other middlewares
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
