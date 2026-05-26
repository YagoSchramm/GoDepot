package derr

var (
	BadRequestError     = NewBadRequestError("Bad Request")
	UnauthorizedError   = NewUnauthorizedError("Unauthorized")
	NotFoundError       = NewNotFoundError("Not Found")
	InternalServerError = NewInternalError("Internal Server Error")
)

var (
	InvalidCredentialsError = NewClientError("INVALID_CREDENTIALS", "Invalid credentials")
	EmailRequired           = NewClientError("EMAIL_REQUIRED", "Email required")
	UserAlreadyExists       = NewClientError("USER_ALREADY_EXISTS", "User already exists")
	PasswordRequired        = NewClientError("PASSWORD_REQUIRED", "Password required")
	InvalidCredentials      = NewClientError("INVALID_CREDENTIALS", "Invalid credentials")
	InvalidEmail            = NewClientError("INVALID_EMAIL", "Invalid email")
)
