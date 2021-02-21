package service

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/ramailh/authentication-server/models"
	"github.com/ramailh/authentication-server/repository"

	"github.com/dgrijalva/jwt-go"
	"github.com/renstrom/shortuuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var confOAuth oauth2.Config

type Services interface {
	Login(username, password string) (map[string]interface{}, error)
	Register(user models.User) (interface{}, error)
	Update(user models.User) (interface{}, error)
	Delete(id string) (interface{}, error)
	FindAll(sortType, sortBy, wid, search string, from, limit int) ([]models.User, error)
	FindByID(id string) (models.User, error)
	Google() string
	GoogleCallback(code string) (map[string]interface{}, error)
	// VerifyToken(token string) bool
}

type service struct {
	repo repository.Repository
}

func NewUserService(repo repository.Repository) Services {
	return &service{repo: repo}
}

func (srv *service) Register(user models.User) (interface{}, error) {
	if user.Password == "" || user.Email == "" || user.Username == "" {
		return nil, errors.New("username, password, and email cannont be empty")
	}

	if srv.repo.DoesUsernameExist(user.Username) {
		return nil, errors.New("username already exists")
	}

	if srv.repo.DoesEmailExist(user.Email) {
		return nil, errors.New("email already exists")
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(password)

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.ID = shortuuid.New()

	return srv.repo.Insert(user)
}

func (srv *service) Google() string {
	confOAuth = oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		RedirectURL:  "http://localhost:8080/login/call-back",
	}

	return confOAuth.AuthCodeURL("state")
}

func (srv *service) GoogleCallback(code string) (map[string]interface{}, error) {
	token, err := confOAuth.Exchange(context.Background(), code)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	client := confOAuth.Client(context.Background(), token)

	responseInfo, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer responseInfo.Body.Close()

	jsonInfo, _ := ioutil.ReadAll(responseInfo.Body)

	data := make(map[string]interface{})
	json.Unmarshal(jsonInfo, &data)

	email := data["email"].(string)

	user, err := srv.repo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	tokenString, err := generateToken(user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"token": tokenString}, nil
}

func (srv *service) Login(username, password string) (map[string]interface{}, error) {
	user, err := srv.repo.FindByUsername(username)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Println(err)
		return nil, err
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS512"), jwt.MapClaims{
		"username":   user.Username,
		"role":       user.Role,
		"timestamp":  time.Now().Unix(),
		"expired_at": time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return map[string]interface{}{"token": tokenString}, nil
}

func (srv *service) Update(user models.User) (interface{}, error) {
	user.UpdatedAt = time.Now()
	return srv.repo.Update(user, user.ID)
}

func (srv *service) Delete(id string) (interface{}, error) {
	return srv.repo.Delete(id)
}

func (srv *service) FindAll(sortType, sortBy, wid, search string, from, limit int) ([]models.User, error) {
	users, err := srv.repo.FindAll(sortBy, sortType, wid, search)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	users = paginate(users, from, limit)

	return users, nil
}

func (srv *service) FindByID(id string) (models.User, error) {
	return srv.repo.FindByID(id)
}

func paginate(users []models.User, from, limit int) []models.User {
	pg := from * limit
	lmtPg := pg + limit

	if len(users) > lmtPg {
		users = users[pg:lmtPg]
	} else if len(users) >= pg {
		users = users[pg:]
	}

	return users
}

func generateToken(username, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS512"), jwt.MapClaims{
		"username":   username,
		"role":       role,
		"timestamp":  time.Now().Unix(),
		"expired_at": time.Now().Add(time.Hour).Unix(),
	})

	return token.SignedString([]byte(os.Getenv("SECRET")))
}
