package chat

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nilspolek/DevOps/Chat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
)

// Mock Storage
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetChats(user uuid.UUID) ([]utils.Chat, error) {
	args := m.Called(user)
	return args.Get(0).([]utils.Chat), args.Error(1)
}

func (m *MockStorage) GetChatMessages(chatId uuid.UUID, limit, offset int) ([]utils.Message, error) {
	args := m.Called(chatId, limit, offset)
	return args.Get(0).([]utils.Message), args.Error(1)
}

func (m *MockStorage) MemberOfChat(userId uuid.UUID, chatId uuid.UUID) error {
	args := m.Called(userId, chatId)
	return args.Error(0)
}

func (m *MockStorage) GetMessage(messageId uuid.UUID) (utils.Message, error) {
	args := m.Called(messageId)
	return args.Get(0).(utils.Message), args.Error(1)
}

func (m *MockStorage) SaveMessage(message utils.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockStorage) CreateOrUpdateChat(chat utils.Chat) error {
	args := m.Called(chat)
	return args.Error(0)
}

func (m *MockStorage) UpdateChatActivity(chat uuid.UUID) error {
	args := m.Called(chat)
	return args.Error(0)
}

func (m *MockStorage) GetChat(id uuid.UUID) (*utils.Chat, error) {
	return nil, nil
}

func (m *MockStorage) DeleteChat(chat uuid.UUID) error {
	args := m.Called(chat)
	return args.Error(0)
}

func (m *MockStorage) UpdateMessage(message utils.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockStorage) DeleteMessage(message uuid.UUID) error {
	args := m.Called(message)
	return args.Error(0)
}

// Mock AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Exists(ids ...uuid.UUID) (bool, error) {
	if len(ids) == 1 {
		args := m.Called(ids[0])
		return args.Bool(0), args.Error(1)
	}
	args := m.Called(ids)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthService) VerifyToken(token string) (*utils.User, error) {
	return nil, nil
}

func TestGetChats(t *testing.T) {
	mockStorage := new(MockStorage)
	mockAuth := new(MockAuthService)
	service := New(mockStorage, mockAuth, nil)

	userId := uuid.New()
	expectedChats := []utils.Chat{
		{
			ID:        uuid.New(),
			Name:      "Test Chat",
			Members:   []uuid.UUID{userId},
			CreatorID: userId,
			CreatedAt: time.Now(),
		},
		{
			ID:       userId,
			Messages: []utils.Message{},
			Members:  []uuid.UUID{userId},
			Name:     "AI",
		}}

	mockStorage.On("GetChats", userId).Return(expectedChats, nil)

	chats, err := service.GetChats(userId)

	assert.NoError(t, err)
	assert.Equal(t, expectedChats, chats)
	mockStorage.AssertExpectations(t)
}

func TestGetMessages(t *testing.T) {
	mockStorage := new(MockStorage)
	mockAuth := new(MockAuthService)
	service := New(mockStorage, mockAuth, nil)

	userId := uuid.New()
	chatId := uuid.New()
	limit := 20
	offset := 0
	expectedMessages := []utils.Message{{
		ID:        uuid.New(),
		Content:   "Test Message",
		SenderID:  userId,
		ChatID:    chatId,
		Timestamp: time.Now(),
	}}

	mockStorage.On("MemberOfChat", userId, chatId).Return(nil)
	mockStorage.On("GetChatMessages", chatId, limit, offset).Return(expectedMessages, nil)

	messages, err := service.GetMessages(userId, chatId, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, expectedMessages, messages)
	mockStorage.AssertExpectations(t)
}

func TestCreateChat(t *testing.T) {
	mockStorage := new(MockStorage)
	mockAuth := new(MockAuthService)
	service := New(mockStorage, mockAuth, nil)

	userId := uuid.New()
	member2 := uuid.New()
	chatName := "Test Chat"
	members := []uuid.UUID{userId, member2}

	mockAuth.On("Exists", members).Return(true, nil)
	mockStorage.On("CreateOrUpdateChat", mock.AnythingOfType("utils.Chat")).Return(nil)

	chat, err := service.CreateChat(userId, chatName, members, nil)

	assert.NoError(t, err)
	assert.Equal(t, chatName, chat.Name)
	assert.Equal(t, members, chat.Members)
	assert.Equal(t, userId, chat.CreatorID)
	mockStorage.AssertExpectations(t)
	mockAuth.AssertExpectations(t)
}

func TestCreateDirectChat(t *testing.T) {
	mockStorage := new(MockStorage)
	mockAuth := new(MockAuthService)
	service := New(mockStorage, mockAuth, nil)

	userId := uuid.New()
	receiverId := uuid.New()
	initialMessage := "Hello!"

	mockAuth.On("Exists", receiverId).Return(true, nil)
	mockStorage.On("CreateOrUpdateChat", mock.AnythingOfType("utils.Chat")).Return(nil)
	mockStorage.On("SaveMessage", mock.AnythingOfType("utils.Message")).Return(nil)

	chat, err := service.CreateDirectChat(userId, receiverId, &initialMessage)

	assert.NoError(t, err)
	assert.Equal(t, "Direct Chat", chat.Name)
	assert.Contains(t, chat.Members, userId)
	assert.Contains(t, chat.Members, receiverId)
	assert.Equal(t, userId, chat.CreatorID)
	mockStorage.AssertExpectations(t)
	mockAuth.AssertExpectations(t)
}

func TestCreateDirectChat_ReceiverDoesNotExist(t *testing.T) {
	mockStorage := new(MockStorage)
	mockAuth := new(MockAuthService)
	service := New(mockStorage, mockAuth, nil)

	userId := uuid.New()
	receiverId := uuid.New()

	mockAuth.On("Exists", receiverId).Return(false, nil)

	_, err := service.CreateDirectChat(userId, receiverId, nil)

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, err.(*utils.ServiceError).StatusCode)
	mockAuth.AssertExpectations(t)
}

func TestGetMessages_WithPagination(t *testing.T) {
	mockStorage := new(MockStorage)
	mockAuth := new(MockAuthService)
	service := New(mockStorage, mockAuth, nil)

	userId := uuid.New()
	chatId := uuid.New()
	limit := 10
	offset := 5

	expectedMessages := []utils.Message{
		{ID: uuid.New(), Content: "Test 1", ChatID: chatId},
		{ID: uuid.New(), Content: "Test 2", ChatID: chatId},
	}

	mockStorage.On("MemberOfChat", userId, chatId).Return(nil)
	mockStorage.On("GetChatMessages", chatId, limit, offset).Return(expectedMessages, nil)

	messages, err := service.GetMessages(userId, chatId, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, expectedMessages, messages)
	mockStorage.AssertExpectations(t)
}

func TestGetMessages_Unauthorized(t *testing.T) {
	mockStorage := new(MockStorage)
	mockAuth := new(MockAuthService)
	service := New(mockStorage, mockAuth, nil)

	userId := uuid.New()
	chatId := uuid.New()

	mockStorage.On("MemberOfChat", userId, chatId).Return(utils.NewError("unauthorized", http.StatusUnauthorized))

	_, err := service.GetMessages(userId, chatId, 10, 0)

	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, err.(*utils.ServiceError).StatusCode)
}

func TestUpdateMessage_Success(t *testing.T) {
	mockStorage := new(MockStorage)
	mockAuth := new(MockAuthService)
	service := New(mockStorage, mockAuth, nil)

	userId := uuid.New()
	messageId := uuid.New()
	originalMessage := utils.Message{
		ID:       messageId,
		SenderID: userId,
		Content:  "original",
	}
	updatedContent := "updated"

	mockStorage.On("GetMessage", messageId).Return(originalMessage, nil)
	mockStorage.On("UpdateMessage", mock.MatchedBy(func(m utils.Message) bool {
		return m.Content == updatedContent && m.ID == messageId
	})).Return(nil)

	result, err := service.UpdateMessage(userId, utils.Message{ID: messageId, Content: updatedContent})

	assert.NoError(t, err)
	assert.Equal(t, updatedContent, result.Content)
}

func TestUpdateMessage_NotSender(t *testing.T) {
	mockStorage := new(MockStorage)
	mockAuth := new(MockAuthService)
	service := New(mockStorage, mockAuth, nil)

	userId := uuid.New()
	differentUserId := uuid.New()
	messageId := uuid.New()
	originalMessage := utils.Message{
		ID:       messageId,
		SenderID: differentUserId,
		Content:  "original",
	}

	mockStorage.On("GetMessage", messageId).Return(originalMessage, nil)

	_, err := service.UpdateMessage(userId, utils.Message{ID: messageId, Content: "updated"})

	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, err.(*utils.ServiceError).StatusCode)
}

func TestDeleteMessage_Success(t *testing.T) {
	mockStorage := new(MockStorage)
	mockAuth := new(MockAuthService)
	service := New(mockStorage, mockAuth, nil)

	userId := uuid.New()
	messageId := uuid.New()
	originalMessage := utils.Message{
		ID:       messageId,
		SenderID: userId,
		Content:  "original",
	}

	mockStorage.On("GetMessage", messageId).Return(originalMessage, nil)
	mockStorage.On("DeleteMessage", messageId).Return(nil)

	result, err := service.DeleteMessage(userId, messageId)

	assert.NoError(t, err)
	assert.True(t, result.Deleted)
	assert.Empty(t, result.Content)
}

func TestDeleteMessage_NotFound(t *testing.T) {
	mockStorage := new(MockStorage)
	mockAuth := new(MockAuthService)
	service := New(mockStorage, mockAuth, nil)

	userId := uuid.New()
	messageId := uuid.New()

	mockStorage.On("GetMessage", messageId).Return(utils.Message{}, mongo.ErrNoDocuments)

	_, err := service.DeleteMessage(userId, messageId)

	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, err.(*utils.ServiceError).StatusCode)
}
