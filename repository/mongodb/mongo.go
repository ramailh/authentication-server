package mongodb

import (
	"context"
	"fmt"
	"github.com/ramailh/authentication-server/models"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	db  = "management"
	col = "user"
)

func ConnectMongo(host, username, password string) (*mongo.Client, error) {
	url := fmt.Sprintf("mongodb://%s/?connect=direct", host)
	opt := options.Client().ApplyURI(url)

	if username != "" && password != "" {
		opt.SetAuth(options.Credential{Username: username, Password: password})
	}

	return mongo.Connect(context.Background(), opt)
}

type repo struct {
	db *mongo.Database
}

func NewMongoRepo(client *mongo.Client) *repo {
	return &repo{db: client.Database(db)}
}

func (rep *repo) FindAll(sortBy, sortType, wid, search string) ([]models.User, error) {
	if sortBy == "" {
		sortBy = "_id"
	}

	if sortType == "" {
		sortType = "desc"
	}

	mapOrder := map[string]int{"asc": 1, "desc": -1}

	opt := options.Find().SetSort(bson.M{sortBy: mapOrder[strings.ToLower(sortType)]}).SetProjection(bson.M{"password": 0})

	var filters []bson.M
	filter := bson.M{}

	if wid != "" {
		filters = append(filters, bson.M{"wid": wid})
	}

	if search != "" {
		pattern := fmt.Sprintf(".*%s.*", search)
		filters = append(filters, bson.M{"username": primitive.Regex{Pattern: pattern, Options: "i"}})
	}

	if len(filters) > 0 {
		filter = bson.M{"$and": filters}
	}

	res, err := rep.db.Collection(col).Find(context.Background(), filter, opt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var users []models.User
	if err = res.All(context.Background(), &users); err != nil {
		log.Println(err)
		return nil, err
	}

	return users, nil
}

func (rep *repo) FindByID(id string) (models.User, error) {
	var user models.User
	err := rep.db.Collection(col).FindOne(context.Background(), bson.M{"_id": id}, options.FindOne().SetProjection(bson.M{"password": 0})).
		Decode(&user)
	return user, err
}

func (rep *repo) FindByEmail(email string) (models.User, error) {
	var user models.User
	err := rep.db.Collection(col).FindOne(context.Background(), bson.M{"email": email}, options.FindOne().SetProjection(bson.M{"password": 0})).
		Decode(&user)
	return user, err
}

func (rep *repo) FindByUsername(username string) (models.User, error) {
	var user models.User
	err := rep.db.Collection(col).FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	return user, err
}

func (rep *repo) FindByUsernameAndPassword(username, password string) (models.User, error) {
	var user models.User
	err := rep.db.Collection(col).
		FindOne(context.Background(), bson.M{"$and": bson.A{bson.M{"username": username}, bson.M{"password": password}}}).
		Decode(&user)
	return user, err
}

func (rep *repo) DoesUsernameExist(username string) bool {
	err := rep.db.Collection(col).FindOne(context.Background(), bson.M{"username": username}).Err()
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (rep *repo) DoesEmailExist(email string) bool {
	err := rep.db.Collection(col).FindOne(context.Background(), bson.M{"email": email}).Err()
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (rep *repo) Insert(doc interface{}) (interface{}, error) {
	return rep.db.Collection(col).InsertOne(context.Background(), doc)
}

func (rep *repo) Update(doc interface{}, id string) (interface{}, error) {
	return rep.db.Collection(col).UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": doc})
}

func (rep *repo) Delete(id string) (interface{}, error) {
	return rep.db.Collection(col).DeleteOne(context.Background(), bson.M{"_id": id})
}

func (rep *repo) Count(wid, search string) (int64, error) {
	var filters []bson.M
	filter := bson.M{}

	if wid != "" {
		filters = append(filters, bson.M{"wid": wid})
	}

	if search != "" {
		pattern := fmt.Sprintf(".*%s.*", search)
		filters = append(filters, bson.M{"username": primitive.Regex{Pattern: pattern, Options: "i"}})
	}

	if len(filters) > 0 {
		filter = bson.M{"$and": filters}
	}

	return rep.db.Collection(col).CountDocuments(context.Background(), filter)
}
