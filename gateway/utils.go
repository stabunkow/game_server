package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"regexp"
)

func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email empty")
	}

	if m, _ := regexp.MatchString(`^([\w\.\_]{2,10})@(\w{1,})\.([a-z]{2,4})$`, email); !m {
		return errors.New("email invalid")
	}

	return nil
}

func ValidatePassword(password string) error {
	if len := len(password); len < 6 || len > 18 {
		return errors.New("password length between 6-18")
	}

	return nil
}

func UniqueId() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
