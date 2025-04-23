package auth

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"team6-managing.mni.thm.de/Commz/auth-service/internal/utils"
)

// MockStorage implements the Storage interface for testing
type MockStorage struct {
	users map[string]utils.User
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		users: make(map[string]utils.User),
	}
}

func (m *MockStorage) UpdateOrCreateUser(user utils.User) error {
	m.users[user.Email] = user
	return nil
}

func (m *MockStorage) GetUserByEmail(email string) (utils.User, error) {
	if user, ok := m.users[email]; ok {
		return user, nil
	}
	return utils.User{}, fmt.Errorf("user not found")
}

func (m *MockStorage) Exists(email string) bool {
	_, exists := m.users[email]
	return exists
}

func (m *MockStorage) GetUserByID(id uuid.UUID) (utils.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return utils.User{}, fmt.Errorf("user not found")
}

func (m *MockStorage) GetUsers() ([]utils.User, error) {
	users := make([]utils.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func TestRegisterUser(t *testing.T) {
	mockStorage := NewMockStorage()
	service := New(mockStorage)

	tests := []struct {
		name          string
		user          utils.User
		expectedError bool
	}{
		{
			name: "Valid user registration",
			user: utils.User{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			expectedError: false,
		},
		{
			name: "Invalid email",
			user: utils.User{
				Email:     "",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			expectedError: true,
		},
		{
			name: "Short password",
			user: utils.User{
				Email:     "test2@example.com",
				Password:  "pass",
				FirstName: "John",
				LastName:  "Doe",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.RegisterUser(tt.user)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, user.ID)
				assert.Equal(t, tt.user.Email, user.Email)
				assert.NotEqual(t, tt.user.Password, user.Password) // Password should be hashed
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	mockStorage := NewMockStorage()
	service := New(mockStorage)

	// Register a test user first
	testUser := utils.User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	registeredUser, err := service.RegisterUser(testUser)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		email         string
		password      string
		expectedError bool
	}{
		{
			name:          "Valid login",
			email:         "test@example.com",
			password:      "password123",
			expectedError: false,
		},
		{
			name:          "Wrong password",
			email:         "test@example.com",
			password:      "wrongpassword",
			expectedError: true,
		},
		{
			name:          "Non-existent user",
			email:         "nonexistent@example.com",
			password:      "password123",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, token, err := service.LoginUser(tt.email, tt.password)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.Equal(t, registeredUser.ID, user.ID)
				assert.Equal(t, registeredUser.Email, user.Email)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	mockStorage := NewMockStorage()
	service := New(mockStorage)

	// Register a test user first
	testUser := utils.User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	registeredUser, err := service.RegisterUser(testUser)
	assert.NoError(t, err)

	t.Run("Get existing user", func(t *testing.T) {
		user, err := service.GetUserByID(registeredUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, registeredUser.ID, user.ID)
		assert.Equal(t, registeredUser.Email, user.Email)
	})

	t.Run("Get non-existent user", func(t *testing.T) {
		_, err := service.GetUserByID(uuid.New())
		assert.Error(t, err)
	})
}

func TestGetUsers(t *testing.T) {
	mockStorage := NewMockStorage()
	service := New(mockStorage)

	// Register multiple test users
	testUsers := []utils.User{
		{
			Email:     "test1@example.com",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
		},
		{
			Email:     "test2@example.com",
			Password:  "password123",
			FirstName: "Jane",
			LastName:  "Smith",
		},
	}

	for _, user := range testUsers {
		_, err := service.RegisterUser(user)
		assert.NoError(t, err)
	}

	t.Run("Get all users", func(t *testing.T) {
		users, err := service.GetUsers()
		assert.NoError(t, err)
		assert.Len(t, users, len(testUsers))
	})
}

func TestUpdateUser(t *testing.T) {
	mockStorage := NewMockStorage()
	service := New(mockStorage)

	// Register a test user first
	testUser := utils.User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	registeredUser, err := service.RegisterUser(testUser)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		userID        uuid.UUID
		updateData    utils.User
		expectedError bool
	}{
		{
			name:   "Valid update",
			userID: registeredUser.ID,
			updateData: utils.User{
				Email:     "updated@example.com",
				FirstName: "Johnny",
				LastName:  "Updated",
			},
			expectedError: false,
		},
		{
			name:   "Invalid email",
			userID: registeredUser.ID,
			updateData: utils.User{
				Email:     "inv",
				FirstName: "Johnny",
				LastName:  "Updated",
			},
			expectedError: true,
		},
		{
			name:   "Non-existent user",
			userID: uuid.New(),
			updateData: utils.User{
				Email:     "test@example.com",
				FirstName: "Johnny",
				LastName:  "Updated",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedUser, err := service.UpdateUser(tt.userID, tt.updateData)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.updateData.Email, updatedUser.Email)
				assert.Equal(t, tt.updateData.FirstName, updatedUser.FirstName)
				assert.Equal(t, tt.updateData.LastName, updatedUser.LastName)
				assert.Equal(t, registeredUser.Password, updatedUser.Password)
			}
		})
	}
}

func TestChangePassword(t *testing.T) {
	mockStorage := NewMockStorage()
	service := New(mockStorage)

	// Register a test user first
	testUser := utils.User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	registeredUser, err := service.RegisterUser(testUser)
	assert.NoError(t, err)

	tests := []struct {
		name            string
		userID          uuid.UUID
		currentPassword string
		newPassword     string
		expectedError   bool
	}{
		{
			name:            "Valid password change",
			userID:          registeredUser.ID,
			currentPassword: "password123",
			newPassword:     "newpassword123",
			expectedError:   false,
		},
		{
			name:            "Wrong current password",
			userID:          registeredUser.ID,
			currentPassword: "wrongpassword",
			newPassword:     "newpassword123",
			expectedError:   true,
		},
		{
			name:            "Invalid new password (too short)",
			userID:          registeredUser.ID,
			currentPassword: "password123",
			newPassword:     "short",
			expectedError:   true,
		},
		{
			name:            "Non-existent user",
			userID:          uuid.New(),
			currentPassword: "password123",
			newPassword:     "newpassword123",
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedUser, err := service.ChangePassword(tt.userID, tt.currentPassword, tt.newPassword)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, registeredUser.Password, updatedUser.Password)

				// Verify the new password works for login
				_, token, loginErr := service.LoginUser(updatedUser.Email, tt.newPassword)
				assert.NoError(t, loginErr)
				assert.NotEmpty(t, token)
			}
		})
	}
}
