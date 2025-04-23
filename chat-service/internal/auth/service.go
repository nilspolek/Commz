package auth

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nilspolek/DevOps/Chat/internal/utils"
)

type ttlUser struct {
	user *utils.User
	time time.Time
}

type AuthService struct {
	gateway string
	cache   map[string]ttlUser
	mu      sync.RWMutex
}

func New(gateway string) AuthService {
	return AuthService{
		gateway: gateway,
		cache:   map[string]ttlUser{},
		mu:      sync.RWMutex{},
	}
}

func (a *AuthService) VerifyToken(token string) (*utils.User, error) {
	a.mu.RLock()
	value, ok := a.cache[token]
	a.mu.RUnlock()

	if ok {
		if time.Since(value.time).Minutes() < 5 {
			return value.user, nil
		}
		a.mu.Lock()
		delete(a.cache, token)
		a.mu.Unlock()
	}

	// get user id from auth service
	request := VerifyTokenRequest{
		Token: token,
	}

	user, err := utils.PostRequest[VerifyTokenRequest, utils.User](a.gateway+"/auth/verify", request)
	if err != nil {
		return nil, err
	}

	a.mu.Lock()
	a.cache[token] = ttlUser{user: user, time: time.Now()}
	a.mu.Unlock()
	return user, nil
}

func (a *AuthService) Exists(ids ...uuid.UUID) (bool, error) {
	users, err := utils.GetRequest[[]utils.User](a.gateway + "/auth/users")
	if err != nil {
		return false, err
	}

	userMap := make(map[uuid.UUID]bool, len(*users))
	for _, user := range *users {
		userMap[user.ID] = true
	}

	for _, id := range ids {
		if !userMap[id] {
			return false, nil
		}
	}
	return true, nil
}
