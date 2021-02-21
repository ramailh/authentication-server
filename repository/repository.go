package repository

import (
	"github.com/ramailh/authentication-server/models"
)

type Repository interface {
	FindAll(sortBy, sortType, wid, search string) ([]models.User, error)
	FindByID(id string) (models.User, error)
	FindByUsernameAndPassword(username, password string) (models.User, error)
	FindByUsername(username string) (models.User, error)
	FindByEmail(email string) (models.User, error)
	DoesUsernameExist(username string) bool
	DoesEmailExist(email string) bool
	Insert(doc interface{}) (interface{}, error)
	Update(doc interface{}, id string) (interface{}, error)
	Delete(id string) (interface{}, error)
	Count(wid, search string) (int64, error)
}
