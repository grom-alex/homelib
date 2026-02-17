package service

import "errors"

var (
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrRegistrationDisabled = errors.New("registration is disabled")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrAccountDeactivated   = errors.New("account is deactivated")
	ErrInvalidRefreshToken  = errors.New("invalid refresh token")
	ErrPasswordTooLong      = errors.New("password too long (max 72 bytes)")
	ErrImportAlreadyRunning = errors.New("import is already running")
)
