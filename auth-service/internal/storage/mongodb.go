package storage

import (
	"context"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"team6-managing.mni.thm.de/Commz/auth-service/internal/utils"
)

const (
	DB_NAME = "commz"
)

type MongoDBStorage struct {
	// MongoDB client
	users *mongo.Collection
}

func NewMongoDBStorage(connectionURI string) (*MongoDBStorage, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connectionURI))
	if err != nil {
		return nil, err
	}
	users := client.Database(DB_NAME).Collection("users")

	_, err = users.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: options.Index().SetUnique(true),
	})

	if err != nil {
		return nil, err
	}

	return &MongoDBStorage{
		users: users,
	}, nil
}

func (m *MongoDBStorage) UpdateOrCreateUser(user utils.User) error {
	ctx := context.Background()
	filter := bson.M{"email": user.Email}
	_, err := m.users.UpdateOne(ctx, filter, bson.M{"$set": user}, options.Update().SetUpsert(true))
	return err
}

func (m *MongoDBStorage) getUser(filter bson.M) (utils.User, error) {
	ctx := context.Background()
	result := m.users.FindOne(ctx, filter)
	if err := result.Err(); err != nil {
		return utils.User{}, err
	}

	var user utils.User
	err := result.Decode(&user)
	if err != nil {
		return utils.User{}, err
	}

	return user, nil
}

func (m *MongoDBStorage) GetUserByEmail(email string) (utils.User, error) {
	filter := bson.M{"email": email}
	return m.getUser(filter)
}

func (m *MongoDBStorage) GetUsers() ([]utils.User, error) {
	ctx := context.Background()
	cursor, err := m.users.Find(ctx, bson.M{}, &options.FindOptions{
		Projection: bson.M{"password": 0},
	})
	if err != nil {
		return nil, err
	}

	var users []utils.User
	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m *MongoDBStorage) GetUserByID(id uuid.UUID) (utils.User, error) {
	filter := bson.M{"_id": id}
	return m.getUser(filter)
}

func (m *MongoDBStorage) Exists(email string) bool {
	ctx := context.Background()
	filter := bson.M{"email": email}
	result := m.users.FindOne(ctx, filter)
	return result.Err() == nil
}
