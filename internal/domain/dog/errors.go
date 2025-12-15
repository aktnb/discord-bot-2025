package dog

import "errors"

var (
	ErrImageNotFound   = errors.New("dog image not found")
	ErrAPIUnavailable  = errors.New("dog API is unavailable")
	ErrInvalidResponse = errors.New("invalid API response")
)
