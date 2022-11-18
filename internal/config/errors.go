package config

import "errors"

var (
	ErrNoUsername = errors.New("missing username")
	ErrNoPassword = errors.New("missing password")
	ErrNoHost     = errors.New("missing host")
	ErrNoPort     = errors.New("missing port")
	ErrNoDatabase = errors.New("missing database")
)
