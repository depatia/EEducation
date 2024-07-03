package clienterrors

import "errors"

var (
	ErrAllFieldsRequired    = errors.New("all fields are required")
	ErrAlreadyExists        = errors.New("user with this email already exists")
	ErrIncorrectCredentials = errors.New("email or password is incorrect")
	ErrUserNotFound         = errors.New("user not found")
	ErrBadJWT               = errors.New("bad JWT")
	ErrBadEmailFormat       = errors.New("bad email format")
)
