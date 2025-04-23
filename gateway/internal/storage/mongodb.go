package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"team6-managing.mni.thm.de/Commz/gateway/internal/utils"
)

const (
	DB_NAME = "commz"
)

type MongoDBStorage struct {
	// MongoDB client
	chatsCollection    *mongo.Collection
	messagesCollection *mongo.Collection
}

func NewMongoDBStorage(connectionURI string) (*MongoDBStorage, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connectionURI))
	if err != nil {
		return nil, err
	}
	chats := client.Database(DB_NAME).Collection("chats")
	messages := client.Database(DB_NAME).Collection("messages")

	// ensure indexes
	_, err = messages.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{"chat_id": 1},
	})
	_, err = messages.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{"updatedAt": -1},
	})
	if err != nil {
		return nil, err
	}

	return &MongoDBStorage{
		chatsCollection:    chats,
		messagesCollection: messages,
	}, nil
}

func (m *MongoDBStorage) GetChat(id []uuid.UUID) ([]utils.Chat, error) {

	filter := bson.M{"_id": bson.M{"$in": id}}

	ctx := context.Background()
	result, err := m.chatsCollection.Find(ctx, filter)

	if err != nil {
		return nil, err
	}

	chat := []utils.Chat{}
	err = result.All(ctx, &chat)
	if err != nil {
		return nil, err
	}

	return chat, nil
}

func (m *MongoDBStorage) GetMessages(time time.Time) ([]utils.Message, error) {
	filter := bson.M{"updatedAt": bson.M{"$gte": time}}
	ctx := context.Background()
	result, err := m.messagesCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	messages := []utils.Message{}
	err = result.All(ctx, &messages)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
