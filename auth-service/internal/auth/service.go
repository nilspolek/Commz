package auth

import (
	"net/http"
	"net/mail"

	"github.com/google/uuid"
	"team6-managing.mni.thm.de/Commz/auth-service/internal/utils"
)

type AuthService struct {
	storage utils.Storage
}

var (
	logger = utils.GetLogger("auth-service")
)

func New(storage utils.Storage) AuthService {
	return AuthService{
		storage: storage,
	}
}

func validateUser(user utils.User) error {
	if user.Email == "" || len(user.Email) < 3 || len(user.Email) > 64 {
		return utils.NewError("email must be between 3 and 64 characters", http.StatusBadRequest)
	}

	if _, err := mail.ParseAddress(user.Email); err != nil {
		return utils.NewError("email address format is invalid", http.StatusBadRequest)
	}

	if user.Password == "" || len(user.Password) < 8 || len(user.Password) > 64 {
		return utils.NewError("password must be between 8 and 64 characters", http.StatusBadRequest)
	}

	if user.FirstName == "" || len(user.FirstName) < 2 || len(user.FirstName) > 64 {
		return utils.NewError("first name must be between 2 and 64 characters", http.StatusBadRequest)
	}

	if user.LastName == "" || len(user.LastName) < 2 || len(user.LastName) > 64 {
		return utils.NewError("last name must be between 2 and 64 characters", http.StatusBadRequest)
	}

	return nil
}

func (a *AuthService) LoginUser(email string, password string) (utils.User, string, error) {

	logger.Info().Str("email", email).Msg("Login user")

	// first check credentials
	user, err := a.storage.GetUserByEmail(email)

	if err != nil {
		return utils.User{}, "", utils.NewError("user not found", http.StatusNotFound)
	}

	if !utils.ComparePasswords(user.Password, password) {
		return utils.User{}, "", utils.NewError("invalid password", http.StatusUnauthorized)
	}

	// generate token
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return utils.User{}, "", err
	}

	logger.Info().Str("email", email).Msg("User logged in")
	return user, token, nil
}

func (a *AuthService) ChangePassword(userId uuid.UUID, currentPassword, newPassword string) (utils.User, error) {

	// get the current user
	user, err := a.storage.GetUserByID(userId)
	if err != nil {
		return utils.User{}, utils.NewError("user not found", http.StatusNotFound)
	}

	logger.Info().Str("email", user.Email).Msg("Updating user password")

	if !utils.ComparePasswords(user.Password, currentPassword) {
		return utils.User{}, utils.NewError("invalid password", http.StatusUnauthorized)
	}

	user.Password = newPassword

	err = validateUser(user)
	if err != nil {
		return utils.User{}, err
	}

	passwordHash, err := utils.HashPassword(user.Password)
	if err != nil {
		return utils.User{}, err
	}

	user.Password = passwordHash

	err = a.storage.UpdateOrCreateUser(user)

	return user, err
}

func (a *AuthService) UpdateUser(userId uuid.UUID, user utils.User) (utils.User, error) {

	logger.Info().Str("email", user.Email).Msg("Updating user")

	if user.Email == "" || len(user.Email) < 3 || len(user.Email) > 64 {
		return utils.User{}, utils.NewError("email must be between 3 and 64 characters", http.StatusBadRequest)
	}

	if _, err := mail.ParseAddress(user.Email); err != nil {
		return utils.User{}, utils.NewError("email address format is invalid", http.StatusBadRequest)
	}

	if user.FirstName == "" || len(user.FirstName) < 2 || len(user.FirstName) > 64 {
		return utils.User{}, utils.NewError("first name must be between 2 and 64 characters", http.StatusBadRequest)
	}

	if user.LastName == "" || len(user.LastName) < 2 || len(user.LastName) > 64 {
		return utils.User{}, utils.NewError("last name must be between 2 and 64 characters", http.StatusBadRequest)
	}

	// get the current user
	current_user, err := a.storage.GetUserByID(userId)
	if err != nil {
		return utils.User{}, utils.NewError("user not found", http.StatusNotFound)
	}

	user.Password = current_user.Password
	user.ID = current_user.ID

	err = a.storage.UpdateOrCreateUser(user)
	if err != nil {
		return utils.User{}, err
	}

	logger.Info().Str("email", user.Email).Str("UUID", user.ID.String()).Msg("User updated")
	return user, nil
}

func (a *AuthService) RegisterUser(user utils.User) (utils.User, error) {

	logger.Info().Str("email", user.Email).Msg("Registering user")

	err := validateUser(user)
	if err != nil {
		return utils.User{}, err
	}

	if a.storage.Exists(user.Email) {
		return utils.User{}, utils.NewError("user already exists", http.StatusConflict)
	}

	// hash password
	password, err := utils.HashPassword(user.Password)
	if err != nil {
		return utils.User{}, err
	}

	user.Password = password
	user.ID = uuid.New()
	err = a.storage.UpdateOrCreateUser(user)
	if err != nil {
		return utils.User{}, err
	}

	logger.Info().Str("email", user.Email).Str("UUID", user.ID.String()).Msg("User registered")
	return user, nil
}

func (a *AuthService) GetUserByID(id uuid.UUID) (utils.User, error) {
	return a.storage.GetUserByID(id)
}

func (a *AuthService) GetUsers() ([]utils.User, error) {
	return a.storage.GetUsers()
}
