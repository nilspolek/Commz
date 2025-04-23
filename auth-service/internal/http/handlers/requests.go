package handlers

// LoginUserRequest represents the login request payload
// @Description Login request model
type LoginUserRequest struct {
	// User's email address
	// @example john.doe@example.com
	Email string `json:"email" example:"john.doe@example.com" validate:"required,email"`
	// User's password
	// @example secretpassword123
	Password string `json:"password" example:"secretpassword123" validate:"required,min=6"`
}

// RegisterUserRequest represents the registration request payload
// @Description Registration request model
type RegisterUserRequest struct {
	// User's email address
	// @example john.doe@example.com
	Email string `json:"email" example:"john.doe@example.com" validate:"required,email"`
	// User's password
	// @example secretpassword123
	Password string `json:"password" example:"secretpassword123" validate:"required,min=6"`
	// User's first name
	// @example John
	FirstName string `json:"first_name" example:"John" validate:"required"`
	// User's last name
	// @example Doe
	LastName string `json:"last_name" example:"Doe" validate:"required"`
}

// UpdateUserRequest represents the update request payload
// @Description Update request model
type UpdateUserRequest struct {
	// User's email address
	// @example john.doe@example.com
	Email string `json:"email" example:"john.doe@example.com" validate:"required,email"`
	// @example John
	FirstName string `json:"first_name" example:"John" validate:"required"`
	// User's last name
	// @example Doe
	LastName string `json:"last_name" example:"Doe" validate:"required"`
	// Picture name. Should be a UUID
	// @example 111111-....-111111
	Picture string `json:"picture"`
}

// ChangePasswordRequest represents the registration request payload
// @Description Registration request model
type ChangePasswordRequest struct {
	// User's password
	// @example secretpassword123
	CurrentPassword string `json:"current_password" example:"secretpassword123" validate:"required,min=6"`
	// User's password
	// @example secretpassword123
	NewPassword string `json:"new_password" example:"secretpassword123" validate:"required,min=6"`
}

// VerifyTokenRequest represents the token verification request payload
// @Description Token verification request model
type VerifyTokenRequest struct {
	// JWT token to verify
	// @example eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." validate:"required"`
}
