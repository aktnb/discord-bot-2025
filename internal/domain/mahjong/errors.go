package mahjong

import "errors"

var (
	ErrImageNotFound   = errors.New("mahjong image not found")
	ErrAPIUnavailable  = errors.New("mahjong API is unavailable")
	ErrInvalidResponse = errors.New("invalid API response")
)
