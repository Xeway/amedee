package model

import "time"

// SerializableCookie is used for session storage of cookies
// (moved from main.go)
type SerializableCookie struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	Expires  time.Time
	Secure   bool
	HttpOnly bool
}
