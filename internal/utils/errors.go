package utils

import "errors"

var (
	ErrURLNotFound       = errors.New("URL not found")
	ErrURLExpired        = errors.New("link has expired")
	ErrCustomCodeExists  = errors.New("custom code already exists, please choose another one")
	ErrInvalidCustomCode = errors.New("custom code length must be between 3 and 20 characters")
	ErrGenerateShortCode = errors.New("failed to generate unique short code after retries")
)