package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"team6-managing.mni.thm.de/Commz/ai-service/internal/utils"
)

type ttlUser struct {
	user *utils.User
	time time.Time
}

type AuthService struct {
	gateway string
	cache   map[string]ttlUser
}

func New(gateway string) AuthService {
	return AuthService{
		gateway: gateway,
		cache:   map[string]ttlUser{},
	}
}

func (a *AuthService) VerifyToken(token string) (*utils.User, error) {

	value, ok := a.cache[token]
	if ok {
		if time.Since(value.time).Minutes() < 5 {
			return value.user, nil
		}
		delete(a.cache, token)
	}

	// get user id from auth service
	request := VerifyTokenRequest{
		Token: token,
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, utils.NewError("failed to marshal verify token request", http.StatusInternalServerError)
	}
	bodyReader := bytes.NewReader(jsonBody)

	response, err := http.Post(a.gateway+"/auth/verify", "application/json", bodyReader)
	if err != nil {
		return nil, utils.NewError("failed to verify token", http.StatusUnauthorized)
	}

	var user utils.User
	err = json.NewDecoder(response.Body).Decode(&user)
	if err != nil {
		return nil, utils.NewError("failed to decode user from auth service", http.StatusInternalServerError)
	}

	a.cache[token] = ttlUser{user: &user, time: time.Now()}
	return &user, nil
}

func (a *AuthService) Exists(ids ...uuid.UUID) (bool, error) {
	// TODO: this is super stupid, we should have a better way to check if a user exists
	response, err := http.Get(a.gateway + "/auth/users")
	if err != nil {
		return false, utils.NewError("failed to get users from auth service", http.StatusInternalServerError)
	}
	var users []utils.User
	err = json.NewDecoder(response.Body).Decode(&users)
	if err != nil {
		return false, utils.NewError("failed to decode users from auth service", http.StatusInternalServerError)
	}

	userMap := make(map[uuid.UUID]bool, len(users))
	for _, user := range users {
		userMap[user.ID] = true
	}

	for _, id := range ids {
		if !userMap[id] {
			return false, nil
		}
	}
	return true, nil
}
