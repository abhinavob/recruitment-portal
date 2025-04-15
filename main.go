package main

import (
	"context"
	"database/sql"
	"gin-app/auth"
	db "gin-app/db"
	sqlc "gin-app/db/sqlc"
	"gin-app/middlewares"
	"log"
	"net/http"
	"os"
	"strings"

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
		userRole := session.Get("role")

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
				"role":  userRole,
			},
		})
	})

	// login routes

	r.GET("/auth/google/login", service.LoginHandler)
	r.GET("/auth/google/login/:role", service.LoginHandler)
	r.GET("/auth/google/callback", service.AuthHandler)
	r.GET("/auth/logout", service.LogoutHandler)

	// applicant routes

	applicantRoutes := r.Group("/applicant")
	applicantRoutes.Use(middlewares.AuthMiddleware(), middlewares.ApplicantOnlyMiddleware())

	applicantRoutes.GET("/dashboard", func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		pictureURL := session.Get("picture").(string)
		jobPosts, err := queries.GetAllJobPosts(context.Background())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		log.Println("pictureURL:", pictureURL)
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":   "Applicant Dashboard",
			"name":    userName,
			"role":    "Applicant",
			"page":    "Dashboard",
			"picture": pictureURL,
			"applicant": gin.H{
				"resume":    "Upload Resume",
				"interview": "Interview Requests",
			},
			"jobPosts": jobPosts,
		})
	})

	applicantRoutes.GET("/profile", func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		pictureURL := session.Get("picture").(string)
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":   "Applicant Profile",
			"name":    userName,
			"role":    "Applicant",
			"page":    "Profile",
			"picture": pictureURL,
			"applicant": gin.H{
				"resume":    "Upload Resume",
				"interview": "Interview Requests",
			},
		})
	})

	applicantRoutes.POST("/profile/update", func(c *gin.Context) {
		session := sessions.Default(c)
		skills := strings.Split(c.PostForm("skills"), ",")
		for i := range skills {
			skills[i] = strings.TrimSpace(skills[i])
		}
		id := session.Get("id").(string)
		uid, err := uuid.Parse(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		queries.UpdateApplicantSkills(context.Background(), sqlc.UpdateApplicantSkillsParams{
			ApplicantID: uid,
			Skills:      skills,
		})
		c.Redirect(http.StatusSeeOther, "/applicant/dashboard")
	})

	applicantRoutes.GET("/upload-resume", func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		pictureURL := session.Get("picture").(string)
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":   "Upload Resume",
			"name":    userName,
			"role":    "Applicant",
			"page":    "Upload Resume",
			"picture": pictureURL,
			"applicant": gin.H{
				"resume":    "Upload Resume",
				"interview": "Interview Requests",
			},
		})
	})

	applicantRoutes.GET("/interview-requests", func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		pictureURL := session.Get("picture").(string)
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":   "Interview Requests",
			"name":    userName,
			"role":    "Applicant",
			"page":    "Interview Requests",
			"picture": pictureURL,
			"applicant": gin.H{
				"resume":    "Upload Resume",
				"interview": "Interview Requests",
			},
		})
	})

	// recruiter routes

	recruiterRoutes := r.Group("/recruiter")
	recruiterRoutes.Use(middlewares.AuthMiddleware(), middlewares.RecruiterOnlyMiddleware())

	r.GET("/recruiter/create-company", func(c *gin.Context) {
		session := sessions.Default(c)
		id := session.Get("id").(string)
		c.HTML(http.StatusOK, "create_company.html", gin.H{
			"ID":    id,
			"Title": "Create Company",
		})
	})

	r.POST("/recruiter/create-company/:id", func(c *gin.Context) {
		session := sessions.Default(c)
		logo := session.Get("logo")
		name := c.PostForm("name")
		description := c.PostForm("description")
		company_id := uuid.New()
		id := c.Param("id")
		uid, err := uuid.Parse(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		companyParams := sqlc.CreateCompanyParams{
			ID:          company_id,
			RecruiterID: uuid.NullUUID{UUID: uid, Valid: true},
			Name:        name,
			Description: sql.NullString{String: description, Valid: true},
			Logo:        sql.NullString{String: logo.(string), Valid: true},
		}
		queries.CreateCompany(context.Background(), companyParams)
		c.Redirect(http.StatusSeeOther, "/recruiter/pending")
	})

	r.POST("/recruiter/create-company/:id/cancel", func(c *gin.Context) {
		id := c.Param("id")
		uid, err := uuid.Parse(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		err = queries.RejectRecruiter(context.Background(), uid)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		service.LogoutHandler(c)
		c.Redirect(http.StatusSeeOther, "/")
	})

	r.GET("/recruiter/pending", middlewares.AuthMiddleware(), func(c *gin.Context) {
		role, _ := c.Get("role")
		if role.(string) != "pending" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Pending recruiters only"})
			return
		}
		c.HTML(http.StatusOK, "pending_page.html", gin.H{
			"Title": "Waiting for Admin Approval",
		})
	})

	recruiterRoutes.GET("/dashboard", func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		pictureURL := session.Get("picture").(string)
		jobPosts, err := queries.GetAllJobPosts(context.Background())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":   "Recruiter Dashboard",
			"name":    userName,
			"role":    "Recruiter",
			"picture": pictureURL,
			"recruiter": gin.H{
				"job":       "Job Posting",
				"interview": "Interview Scheduling",
				"resume":    "Resume Parsing",
			},
			"page":     "Dashboard",
			"jobPosts": jobPosts,
		})
	})

	recruiterRoutes.GET("/job-posting", func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		pictureURL := session.Get("picture").(string)
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":   "Recruiter Dashboard",
			"name":    userName,
			"role":    "Recruiter",
			"picture": pictureURL,
			"recruiter": gin.H{
				"job":       "Job Posting",
				"interview": "Interview Scheduling",
				"resume":    "Resume Parsing",
			},
			"page": "Job Posting",
		})
	})

	recruiterRoutes.POST("/job-posting/create", func(c *gin.Context) {
		session := sessions.Default(c)
		company_name := c.PostForm("company_name")
		position := c.PostForm("position")
		description := c.PostForm("description")
		salary := c.PostForm("salary")
		skills := strings.Split(c.PostForm("skills"), ",")
		for i := range skills {
			skills[i] = strings.TrimSpace(skills[i])
		}
		jobID := uuid.New()
		id := session.Get("id").(string)
		uid, err := uuid.Parse(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		company, err := queries.GetCompanyByRecruiterID(context.Background(), uuid.NullUUID{UUID: uid, Valid: true})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		jobPostParams := sqlc.CreateJobPostParams{
			ID:          jobID,
			RecruiterID: uuid.NullUUID{UUID: uid, Valid: true},
			CompanyID:   uuid.NullUUID{UUID: company.ID, Valid: true},
			CompanyName: company_name,
			Position:    position,
			Skills:      skills,
			Description: sql.NullString{String: description, Valid: true},
			Salary:      sql.NullString{String: salary, Valid: true},
		}
		err = queries.CreateJobPost(context.Background(), jobPostParams)
		if err != nil {
			log.Println("error:", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Redirect(http.StatusSeeOther, "/recruiter/dashboard")
	})

	recruiterRoutes.POST("/job-posting/delete/:id", func(c *gin.Context) {
		id := c.Param("id")
		uid, err := uuid.Parse(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		err = queries.DeleteJobPost(context.Background(), uid)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Redirect(http.StatusSeeOther, "/recruiter/dashboard")
	})

	recruiterRoutes.GET("/interview-scheduling", func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		pictureURL := session.Get("picture").(string)
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":   "Recruiter Dashboard",
			"name":    userName,
			"role":    "Recruiter",
			"picture": pictureURL,
			"recruiter": gin.H{
				"job":       "Job Posting",
				"interview": "Interview Scheduling",
				"resume":    "Resume Parsing",
			},
			"page": "Interview Scheduling",
		})
	})

	recruiterRoutes.GET("/resume-parsing", func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		pictureURL := session.Get("picture").(string)
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":   "Recruiter Dashboard",
			"name":    userName,
			"role":    "Recruiter",
			"picture": pictureURL,
			"recruiter": gin.H{
				"job":       "Job Posting",
				"interview": "Interview Scheduling",
				"resume":    "Resume Parsing",
			},
			"page": "Resume Parsing",
		})
	})

	// admin routes

	adminRoutes := r.Group("/admin")
	adminRoutes.Use(middlewares.AuthMiddleware(), middlewares.AdminOnlyMiddleware())

	adminRoutes.GET("/dashboard", func(c *gin.Context) {
		session := sessions.Default(c)
		userName := session.Get("name")
		pictureURL := session.Get("picture").(string)
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":   "Admin Dashboard",
			"name":    userName,
			"role":    "Admin",
			"picture": pictureURL,
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
		pictureURL := session.Get("picture").(string)
		users, err := queries.GetAllUsers(context.Background())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":   "View Users",
			"name":    userName,
			"role":    "Admin",
			"picture": pictureURL,
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
		pictureURL := session.Get("picture").(string)
		pendingRecruiters, err := queries.GetPendingRecruiters(context.Background())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":   "Pending Recruiters",
			"name":    userName,
			"role":    "Admin",
			"picture": pictureURL,
			"admin": gin.H{
				"view": "View Users",
				"add":  "Pending Recruiters",
			},
			"page":              "Pending Recruiters",
			"pendingRecruiters": pendingRecruiters,
		})
	})

	adminRoutes.POST("/approve-recruiter/:id", func(c *gin.Context) {
		id := c.Param("id")
		uid, err := uuid.Parse(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		err = queries.ApproveRecruiter(context.Background(), uid)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Redirect(http.StatusSeeOther, "/admin/pending-recruiters")
	})

	adminRoutes.POST("/reject-recruiter/:id", func(c *gin.Context) {
		id := c.Param("id")
		uid, err := uuid.Parse(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		err = queries.RejectRecruiter(context.Background(), uid)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = queries.RejectCompany(context.Background(), uuid.NullUUID{UUID: uid, Valid: true})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Redirect(http.StatusSeeOther, "/admin/pending-recruiters")
	})

	r.Run(":8080")
}
