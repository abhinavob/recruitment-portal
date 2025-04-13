package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	db "gin-app/db/sqlc"
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

var (
	conf *oauth2.Config
)

type Service struct {
	Queries *db.Queries
}

func NewService(queries *db.Queries) *Service {
	return &Service{Queries: queries}
}

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

func (s *Service) LoginHandler(c *gin.Context) {
	state := randToken()
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()
	c.Redirect(http.StatusFound, conf.AuthCodeURL(state))
}

func (s *Service) AuthHandler(c *gin.Context) {
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

	createdUser, err := s.Queries.GetUserByEmail(context.Background(), user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			params := db.CreateUserParams{
				Name:    user.Name,
				Email:   user.Email,
				Picture: sql.NullString{String: user.Picture, Valid: true},
			}
			createdUser, err = s.Queries.CreateUser(context.Background(), params)
			if err != nil {
				log.Printf("Error creating user: %v", err)
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
		} else {
			log.Printf("Error getting user: %v", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	session.Set("email", createdUser.Email)
	session.Set("name", createdUser.Name)
	if err := session.Save(); err != nil {
		log.Printf("Error saving session: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, os.Getenv("LOGIN_REDIRECT_URL"))
}

func (s *Service) LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)

	session.Clear()

	if err := session.Save(); err != nil {
		log.Printf("Error saving session during logout: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, os.Getenv("LOGOUT_REDIRECT_URL"))
}
