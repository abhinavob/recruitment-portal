package main

import (
	"gin-app/auth"
	db "gin-app/db"
	sqlc "gin-app/db/sqlc"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
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

	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		if session == nil {
			log.Println("Session is nil!")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		userEmail := session.Get("email")
		userName := session.Get("name")

		if userEmail == nil {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"title": "Recruitment Portal",
				"user":  nil,
			})
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Recruitment Portal",
			"user": gin.H{
				"email": userEmail,
				"name":  userName,
			},
		})
	})

	r.GET("/auth/google/login", service.LoginHandler)
	r.GET("/auth/google/callback", service.AuthHandler)
	r.GET("/auth/logout", service.LogoutHandler)

	r.Run(":8080")
}
