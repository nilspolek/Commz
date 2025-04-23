package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"team6-managing.mni.thm.de/Commz/auth-service/internal/auth"
	"team6-managing.mni.thm.de/Commz/auth-service/internal/utils"
)

const (
	NoBodyError = "no body found"
	NoAuthToken = "no auth token found"
)

type AuthHandler struct {
	auth *auth.AuthService
}

func (c *AuthHandler) RegisterRoutes(router *mux.Router) {

	// Auth endpoints
	HandleFunc(router, "/version", c.getVersion, "GET")

	HandleFunc(router, "/user", c.getUser, "GET")
	HandleFunc(router, "/user", c.updateUser, "PUT")
	HandleFunc(router, "/user/password", c.updatePassword, "PUT")

	HandleFunc(router, "/users", c.getUsers, "GET")

	HandleFunc(router, "/register", c.registerUser, "POST")
	HandleFunc(router, "/login", c.loginUser, "POST")

	HandleFunc(router, "/refresh", c.loginUser, "POST") // token refresh is same as login for now
	HandleFunc(router, "/verify", c.verifyToken, "POST")
	HandleFunc(router, "/logout", c.logout, "GEt")
}

// @Summary Get the service Version
// @Success 200 {object} string "version"
// @Router /version [get]
func (c *AuthHandler) getVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(utils.VERSION))
}

// @Summary Logs out a user
// @Description Removes the commz token from the user
// @Tags auth
// @Success 200 {object} bool "User logged out"
// @Router /logout [get]
func (c *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	// set token in cookie
	http.SetCookie(w, &http.Cookie{
		Name:    utils.CommzToken,
		Value:   "",
		Expires: time.Now(),
		Secure:  false,
		Path:    "/",
	})

	utils.SendJsonResponse(w, true)
}

// @Summary Change users password
// @Description Change a users password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "User registration details"
// @Success 200 {object} utils.User "Created user information"
// @Failure 400 {object} utils.ServiceError "Invalid request body"
// @Failure 409 {object} utils.ServiceError "Email already exists"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /user/password [put]
func (c *AuthHandler) updatePassword(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		c.error(w, NoBodyError, http.StatusBadRequest)
		return
	}

	// parse request
	var update ChangePasswordRequest
	err = json.Unmarshal(b, &update)
	if err != nil {
		c.error(w, "invalid update request", http.StatusBadRequest)
		return
	}

	cookies := r.CookiesNamed(utils.CommzToken)
	if len(cookies) == 0 {
		logger.Warn().Msg(NoAuthToken)
		c.error(w, NoAuthToken, http.StatusUnauthorized)
		return
	}

	token := cookies[0].Value
	userId, err := utils.VerifyJWT(token)

	if err != nil {
		c.handleErrors(err, w)
		return
	}

	// this has to be valid because it was verified by the VerifyJWT method
	userUUID := uuid.MustParse(userId)
	// register user
	userResponse, err := c.auth.ChangePassword(userUUID, update.CurrentPassword, update.NewPassword)

	if c.handleErrors(err, w) {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    utils.CommzToken,
		Value:   token,
		Expires: utils.GetTokenExpiry(),
		Secure:  false,
		Path:    "/",
	})

	// send user as response
	utils.SendJsonResponse(w, userResponse)
}

// @Summary Update user
// @Description Updates a user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body UpdateUserRequest true "User registration details"
// @Success 200 {object} utils.User "Created user information"
// @Failure 400 {object} utils.ServiceError "Invalid request body"
// @Failure 409 {object} utils.ServiceError "Email already exists"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /user [put]
func (c *AuthHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		c.error(w, NoBodyError, http.StatusBadRequest)
		return
	}

	// parse request
	var update UpdateUserRequest
	err = json.Unmarshal(b, &update)
	if err != nil {
		c.error(w, "invalid update request", http.StatusBadRequest)
		return
	}

	cookies := r.CookiesNamed(utils.CommzToken)
	if len(cookies) == 0 {
		logger.Warn().Msg(NoAuthToken)
		c.error(w, NoAuthToken, http.StatusUnauthorized)
		return
	}

	token := cookies[0].Value
	userId, err := utils.VerifyJWT(token)

	if err != nil {
		c.handleErrors(err, w)
		return
	}

	// this has to be valid because it was verified by the VerifyJWT method
	userUUID := uuid.MustParse(userId)

	user := utils.User{
		Email:     update.Email,
		FirstName: update.FirstName,
		LastName:  update.LastName,
		Picture:   update.Picture,
	}

	// register user
	userResponse, err := c.auth.UpdateUser(userUUID, user)

	if c.handleErrors(err, w) {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    utils.CommzToken,
		Value:   token,
		Expires: utils.GetTokenExpiry(),
		Secure:  false,
		Path:    "/",
	})

	// send user as response
	utils.SendJsonResponse(w, userResponse)
}

// @Summary Verify JWT token
// @Description Validates a JWT token and returns the associated user information
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyTokenRequest true "Token verification request"
// @Success 200 {object} utils.User "User information"
// @Failure 400 {object} utils.ServiceError "Invalid request body or missing token"
// @Failure 401 {object} utils.ServiceError "Invalid or expired token"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /verify [post]
func (c *AuthHandler) verifyToken(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		c.error(w, NoBodyError, http.StatusBadRequest)
		return
	}

	// parse request
	var verify VerifyTokenRequest
	err = json.Unmarshal(b, &verify)
	if err != nil {
		c.error(w, "invalid verify request", http.StatusBadRequest)
		return
	}

	// register user
	userIDString, err := utils.VerifyJWT(verify.Token)
	if err != nil {
		err = utils.NewError("invalid token", http.StatusUnauthorized)
		c.handleErrors(err, w)
		return
	}

	userID := uuid.MustParse(userIDString)
	userResponse, err := c.auth.GetUserByID(userID)
	if c.handleErrors(err, w) {
		return
	}

	// send user as response
	utils.SendJsonResponse(w, userResponse)
}

// @Summary Login user
// @Description Authenticates a user and returns a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginUserRequest true "Login credentials"
// @Success 200 {object} utils.User "User information with JWT token in cookie"
// @Failure 400 {object} utils.ServiceError "Invalid request body"
// @Failure 401 {object} utils.ServiceError "Invalid credentials"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /login [post]
func (c *AuthHandler) loginUser(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		c.error(w, NoBodyError, http.StatusBadRequest)
		return
	}

	// parse request
	var login LoginUserRequest
	err = json.Unmarshal(b, &login)
	if err != nil {
		c.error(w, "invalid login request", http.StatusBadRequest)
		return
	}

	// register user
	userResponse, token, err := c.auth.LoginUser(login.Email, login.Password)

	if c.handleErrors(err, w) {
		return
	}

	// set token in cookie
	http.SetCookie(w, &http.Cookie{
		Name:    utils.CommzToken,
		Value:   token,
		Expires: utils.GetTokenExpiry(),
		Secure:  false,
		Path:    "/",
	})

	// send user as response
	utils.SendJsonResponse(w, userResponse)
}

// @Summary Register new user
// @Description Creates a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterUserRequest true "User registration details"
// @Success 200 {object} utils.User "Created user information"
// @Failure 400 {object} utils.ServiceError "Invalid request body"
// @Failure 409 {object} utils.ServiceError "Email already exists"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /register [post]
func (c *AuthHandler) registerUser(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		c.error(w, NoBodyError, http.StatusBadRequest)
		return
	}

	// parse request
	var register RegisterUserRequest
	err = json.Unmarshal(b, &register)
	if err != nil {
		c.error(w, "invalid register request", http.StatusBadRequest)
		return
	}

	user := utils.User{
		Email:     register.Email,
		Password:  register.Password,
		FirstName: register.FirstName,
		LastName:  register.LastName,
	}

	// register user
	userResponse, err := c.auth.RegisterUser(user)

	if c.handleErrors(err, w) {
		return
	}

	_, token, err := c.auth.LoginUser(user.Email, user.Password)
	if c.handleErrors(err, w) {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    utils.CommzToken,
		Value:   token,
		Expires: utils.GetTokenExpiry(),
		Secure:  false,
		Path:    "/",
	})

	// send user as response
	utils.SendJsonResponse(w, userResponse)
}

// @Summary Get all users
// @Description Retrieves a list of all registered users
// @Tags auth
// @Produce json
// @Success 200 {array} utils.User "List of users"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /users [get]
func (c *AuthHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := c.auth.GetUsers()
	if c.handleErrors(err, w) {
		return
	}
	utils.SendJsonResponse(w, users)
}

// @Summary Get user
// @Description Retrieves the currently logged in user
// @Tags auth
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Success 200 {object} utils.User "Logged in user"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /users [get]
func (c *AuthHandler) getUser(w http.ResponseWriter, r *http.Request) {
	cookies := r.CookiesNamed(utils.CommzToken)
	if len(cookies) == 0 {
		logger.Warn().Msg(NoAuthToken)
		c.error(w, NoAuthToken, http.StatusUnauthorized)
		return
	}

	token := cookies[0].Value

	userId, err := utils.VerifyJWT(token)

	if err != nil {
		c.handleErrors(err, w)
		return
	}

	// this has to be valid because it was verified by the VerifyJWT method
	userUUID := uuid.MustParse(userId)
	user, err := c.auth.GetUserByID(userUUID)

	if err != nil {
		c.handleErrors(err, w)
		return
	}

	utils.SendJsonResponse(w, user)
}

func (c *AuthHandler) handleErrors(err error, w http.ResponseWriter) bool {
	// check if error is a custom error
	customError, ok := err.(*utils.ServiceError)
	if ok {
		// marshal error into json and send status from error as statuscode
		w.WriteHeader(customError.StatusCode)
		w.Write(customError.Bytes())
		return true
	}

	if err != nil {
		c.error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	return false
}

func (c *AuthHandler) error(w http.ResponseWriter, error string, code int) {
	err := utils.NewError(error, code)
	c.handleErrors(err, w)
}
