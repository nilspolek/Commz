package ws

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/google/uuid"
	"team6-managing.mni.thm.de/Commz/gateway/internal/storage"
)

const (
	UPDATE_TIME = 1 * time.Second
)

type MessageCrawler struct {
	lastUpdate time.Time
	hub        *Hub
	storage    *storage.MongoDBStorage
}

func NewCrawler(databaseUrl string, hub *Hub) (MessageCrawler, error) {

	db, err := storage.NewMongoDBStorage(databaseUrl)

	if err != nil {
		return MessageCrawler{}, err
	}

	return MessageCrawler{
		storage:    db,
		hub:        hub,
		lastUpdate: time.Now(),
	}, nil
}

func (m *MessageCrawler) Run() {
	for {
		time.Sleep(UPDATE_TIME)
		messages, err := m.storage.GetMessages(m.lastUpdate)
		if err != nil {
			logger.Err(err).Msg("error while fetching the latest messages")
			continue
		}
		chatIds := []uuid.UUID{}
		for _, message := range messages {
			chatIds = append(chatIds, message.ChatID)
		}
		chats, err := m.storage.GetChat(chatIds)
		if err != nil {
			logger.Err(err).Msg("error while fetching the latest chats")
			continue
		}

		chatMap := map[uuid.UUID][]uuid.UUID{}

		for _, chat := range chats {
			chatMap[chat.ID] = chat.Members
		}

		slices.Reverse(messages)

		for _, message := range messages {

			bytes, err := json.Marshal(message)
			if err != nil {
				logger.Err(err).Msg("error while marshaling message")
				continue
			}

			boradcastMsg := BroadCastMessage{
				Bytes:    bytes,
				Receiver: chatMap[message.ChatID],
			}

			logger.Debug().
				Str("message", string(bytes)).
				Interface("receivers", boradcastMsg.Receiver).
				Msg("broadcasting message")
			m.hub.broadcast <- boradcastMsg
		}

		m.lastUpdate = time.Now()
	}
}
