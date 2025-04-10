package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type User struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Gender        string `json:"gender"`
}

var conf *oauth2.Config

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err.Error())
		log.Fatal("Error loading .env file")
	}

	conf = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func LoginHandler(c *gin.Context) {
	state := randToken()
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()
	c.Redirect(http.StatusFound, conf.AuthCodeURL(state))
}

func AuthHandler(c *gin.Context) {
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid session state: %s", retrievedState))
		return
	}

	tok, err := conf.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		log.Printf("Error exchanging code: %v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client := conf.Client(context.Background(), tok)
	email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer email.Body.Close()

	data, err := io.ReadAll(email.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		log.Printf("Error unmarshaling user data: %v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	session.Set("user", user.Email)
	session.Save()

	c.Redirect(http.StatusFound, "/")
}
