package serviceerrors

import "errors"

var (
	ErrAccessDenied       = errors.New("Not enough permissions")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrBadEmailFormat     = errors.New("bad email format")
	ErrClassNotFound      = errors.New("class not found")
)
