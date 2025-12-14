package cat

import "errors"

var (
	ErrImageNotFound   = errors.New("cat image not found")
	ErrAPIUnavailable  = errors.New("cat API is unavailable")
	ErrInvalidResponse = errors.New("invalid API response")
)
