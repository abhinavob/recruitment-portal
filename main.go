package main

import (
	"context"
	"gin-app/auth"
	db "gin-app/db"
	sqlc "gin-app/db/sqlc"
	"gin-app/middlewares"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	DB := db.NewDB()
	queries := sqlc.New(DB)
	service := auth.NewService(queries)

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	secret := os.Getenv("SESSION_SECRET")
	store := cookie.NewStore([]byte(secret))
	store.Options(sessions.Options{
		MaxAge: 86400,
		Path:   "/",
	})
	r.Use(sessions.Sessions("mysession", store))

	r.Static("/static", "./static")

	r.LoadHTMLGlob("templates/**/*")

	r.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		userEmail := session.Get("email")
		userName := session.Get("name")

		if userEmail == nil {
			c.HTML(http.StatusOK, "home.html", gin.H{
				"title": "Recruitment Portal",
				"user":  nil,
			})
			return
		}

		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "Recruitment Portal",
			"user": gin.H{
				"email": userEmail,
				"name":  userName,
			},
		})
	})

	// login stuff
	r.GET("/auth/google/login", service.LoginHandler)
	r.GET("/auth/google/login/:role", service.LoginHandler)
	r.GET("/auth/google/callback", service.AuthHandler)
	r.GET("/auth/logout", service.LogoutHandler)

	// applicant stuff
	r.GET("/applicant/dashboard", middlewares.AuthMiddleware(), middlewares.ApplicantOnlyMiddleware(), func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title": "Applicant Dashboard",
			"name":  userName,
			"role":  "Applicant",
		})
	})

	// recruiter stuff
	r.GET("/recruiter/dashboard", middlewares.AuthMiddleware(), middlewares.RecruiterOnlyMiddleware(), func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title": "Recruiter Dashboard",
			"name":  userName,
			"role":  "Recruiter",
		})
	})

	r.GET("/recruiter/pending", middlewares.AuthMiddleware(), func(c *gin.Context) {
		role, _ := c.Get("role")
		if role.(string) != "pending" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Pending users only"})
			return
		}
		c.HTML(http.StatusOK, "pending_page.html", gin.H{})
	})

	// admin stuff
	adminRoutes := r.Group("/admin")
	adminRoutes.Use(middlewares.AuthMiddleware(), middlewares.AdminOnlyMiddleware())
	adminRoutes.GET("/dashboard", func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title": "Admin Dashboard",
			"name":  userName,
			"role":  "Admin",
			"admin": gin.H{
				"view": "View Users",
				"add":  "Pending Recruiters",
			},
			"page": "Dashboard",
		})
	})

	adminRoutes.GET("/view-users", func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		users, err := queries.GetAllUsers(context.Background())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title": "View Users",
			"name":  userName,
			"role":  "Admin",
			"admin": gin.H{
				"view": "View Users",
				"add":  "Pending Recruiters",
			},
			"page":  "View Users",
			"users": users,
		})
	})

	adminRoutes.GET("/pending-recruiters", func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		pendingRecruiters, err := queries.GetPendingRecruiters(context.Background())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title": "Pending Recruiters",
			"name":  userName,
			"role":  "Admin",
			"admin": gin.H{
				"view": "View Users",
				"add":  "Pending Recruiters",
			},
			"page":              "Pending Recruiters",
			"pendingRecruiters": pendingRecruiters,
		})
	})

	adminRoutes.POST("/approve-recruiter/:id", func(c *gin.Context) {
		log.Println("APPROVE TEST")
		id := c.Param("id")
		uuid, err := uuid.Parse(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		err = queries.ApproveRecruiter(context.Background(), uuid)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Redirect(http.StatusSeeOther, "/admin/pending-recruiters")
	})

	adminRoutes.POST("/reject-recruiter/:id", func(c *gin.Context) {
		log.Println("REJECT TEST")
		id := c.Param("id")
		uuid, err := uuid.Parse(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		err = queries.RejectRecruiter(context.Background(), uuid)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Redirect(http.StatusSeeOther, "/admin/pending-recruiters")
	})

	r.Run(":8080")
}
