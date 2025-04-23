package storage

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/nilspolek/DevOps/Chat/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	if err != nil {
		return nil, err
	}

	return &MongoDBStorage{
		chatsCollection:    chats,
		messagesCollection: messages,
	}, nil
}

func (m *MongoDBStorage) GetChat(id uuid.UUID) (*utils.Chat, error) {

	filter := bson.M{"_id": id}

	ctx := context.Background()
	result := m.chatsCollection.FindOne(ctx, filter)

	chat := utils.Chat{}
	err := result.Decode(&chat)
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (m *MongoDBStorage) GetChats(user uuid.UUID) ([]utils.Chat, error) {

	filter := bson.M{"members": user}

	ctx := context.Background()
	result, err := m.chatsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	chats := []utils.Chat{}
	err = result.All(ctx, &chats)

	if err != nil {
		return nil, err
	}

	// fetch the last 10 messages that were sent in this chat
	for i, chat := range chats {
		filter := bson.M{"chat_id": chat.ID}
		opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(30)
		msgResult, err := m.messagesCollection.Find(ctx, filter, opts)
		if err != nil {
			return nil, err
		}

		messages := []utils.Message{}
		if err = msgResult.All(ctx, &messages); err != nil {
			return nil, err
		}

		// remove content of deleted messages
		for i := range messages {
			if messages[i].Deleted {
				messages[i].Content = ""
			}
		}

		slices.Reverse(messages)
		chats[i].Messages = messages
	}

	return chats, nil
}

func (m *MongoDBStorage) GetChatMessages(chatId uuid.UUID, limit, offset int) ([]utils.Message, error) {
	filter := bson.M{"chat_id": chatId}
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	ctx := context.Background()
	result, err := m.messagesCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	messages := []utils.Message{}
	err = result.All(ctx, &messages)
	if err != nil {
		return nil, err
	}

	slices.Reverse(messages)
	return messages, nil
}

func (m *MongoDBStorage) MemberOfChat(userId uuid.UUID, chatId uuid.UUID) error {
	filter := bson.M{"_id": chatId, "members": userId}
	ctx := context.Background()
	result := m.chatsCollection.FindOne(ctx, filter)
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

func (m *MongoDBStorage) SaveMessage(message utils.Message) error {
	ctx := context.Background()
	_, err := m.messagesCollection.InsertOne(ctx, message)
	return err
}

func (m *MongoDBStorage) CreateOrUpdateChat(chat utils.Chat) error {
	ctx := context.Background()
	filter := bson.M{"_id": chat.ID}
	_, err := m.chatsCollection.UpdateOne(ctx, filter, bson.M{"$set": chat}, options.Update().SetUpsert(true))
	return err
}

func (m *MongoDBStorage) DeleteChat(chat uuid.UUID) error {
	ctx := context.Background()
	filter := bson.M{"_id": chat}
	_, err := m.chatsCollection.DeleteOne(ctx, filter)
	return err
}

func (m *MongoDBStorage) DeleteMessage(message uuid.UUID) error {
	ctx := context.Background()
	filter := bson.M{"_id": message}
	_, err := m.messagesCollection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"deleted": true, "updatedAt": time.Now()}})
	return err
}

func (m *MongoDBStorage) UpdateMessage(message utils.Message) error {
	ctx := context.Background()
	filter := bson.M{"_id": message.ID}
	_, err := m.messagesCollection.UpdateOne(ctx, filter, bson.M{"$set": message}, options.Update().SetUpsert(true))
	return err
}

func (m *MongoDBStorage) UpdateChatActivity(chat uuid.UUID) error {
	ctx := context.Background()
	filter := bson.M{"_id": chat}
	_, err := m.chatsCollection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"last_active": time.Now()}})
	return err
}

func (m *MongoDBStorage) GetMessage(messageId uuid.UUID) (utils.Message, error) {
	filter := bson.M{"_id": messageId}
	ctx := context.Background()
	result := m.messagesCollection.FindOne(ctx, filter)
	if result.Err() != nil {
		return utils.Message{}, result.Err()
	}
	message := utils.Message{}
	err := result.Decode(&message)
	if err != nil {
		return utils.Message{}, err
	}
	if message.Deleted {
		message.Content = ""
	}
	return message, nil
}
