package model

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrForbidden             = errors.New("forbidden")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrBonusRemarksRequired  = errors.New("remarks are required when awarding bonus marks")
)
