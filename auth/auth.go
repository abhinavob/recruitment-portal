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
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	role := c.Param("role")
	session.Set("role", role)

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
			role := session.Get("role")
			if role == "recruiter" {
				role = "pending"
			} else if role == "admin" {
				c.Redirect(http.StatusFound, "/")
				return
			}

			params := db.CreateUserParams{
				Name:    user.Name,
				Email:   user.Email,
				Picture: sql.NullString{String: user.Picture, Valid: true},
				Role:    role.(string),
			}
			createdUser, err = s.Queries.CreateUser(context.Background(), params)
			if err != nil {
				log.Printf("Error creating user: %v", err)
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			if role == "pending" {
				session.Set("logo", createdUser.Picture.String)
			}
		} else {
			log.Printf("Error getting user: %v", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	rand_tok := randToken()
	_, err = s.Queries.CreateOrUpdateSession(context.Background(), db.CreateOrUpdateSessionParams{
		UserID: uuid.NullUUID{UUID: createdUser.ID, Valid: true},
		Token:  rand_tok,
	})
	if err != nil {
		log.Printf("Error creating/updating session: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	session.Set("email", createdUser.Email)
	session.Set("name", createdUser.Name)
	session.Set("role", createdUser.Role)
	session.Set("id", createdUser.ID.String())
	session.Set("picture", createdUser.Picture.String)
	session.Set("token", rand_tok)
	if err := session.Save(); err != nil {
		log.Printf("Error saving session: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, os.Getenv(strings.ToUpper(createdUser.Role)+"_REDIRECT_URL"))
}

func (s *Service) LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)

	token := session.Get("token")
	email := session.Get("email")

	if token != nil && email != nil {
		user, err := s.Queries.GetUserByEmail(context.Background(), email.(string))
		if err != nil {
			log.Printf("Error getting user during logout: %v", err)
		} else {
			err = s.Queries.DeleteSession(context.Background(), db.DeleteSessionParams{
				UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
				Token:  token.(string),
			})
			if err != nil {
				log.Printf("Error deleting session: %v", err)
			}
		}
	}

	session.Clear()
	if err := session.Save(); err != nil {
		log.Printf("Error saving session during logout: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, os.Getenv("LOGOUT_REDIRECT_URL"))
}
