package main

import (
	"net/http"

	"gin-app/auth"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	store := cookie.NewStore([]byte("secret-key"))
	r.Use(sessions.Sessions("mysession", store))

	r.Static("/static", "./static")

	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Welcome!",
		})
	})

	r.GET("/auth/google/login", auth.LoginHandler)
	r.GET("/auth/google/callback", auth.AuthHandler)

	r.Run(":8080")
}
