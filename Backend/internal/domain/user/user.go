package user

import (
	"errors"
	"net/mail"
	"strings"
)

type User struct {
	Name     string
	Email    string
	Password string
}

// errors with full information handle on client
var domainErr = func(field string) error { return errors.New("payload validation error: from field '" + field + "'") }

var (
	IncorrectEmailAddrErr error = errors.New("domain error: invalid email")
	IncorrectPasswordErr  error = errors.New("domain error: not equal password")
)

func NewUser(name, email, password string) (User, error) {
	if err := validate(name, "username"); err != nil {
		return User{}, err
	} else if err = validate(email, "email"); err != nil {
		return User{}, err
	}
	name = strings.ReplaceAll(name, " ", "_")
	// optional: validate password (in future)
	return User{Name: name, Email: email, Password: password}, nil
}

func validate(val, field string) error {
	switch field {
	case "name":
		if len([]byte(val)) > 8 ||
			len([]byte(val)) < 3 ||
			strings.ContainsAny(val, "~!@#$%^&*()_+={}[]:;\"\\/'`-?№<>.,|1234567890") {
			return domainErr(field)
		}
	case "email":
		if _, err := mail.ParseAddress(val); err != nil {
			return domainErr(field)
		}

		// password need check from client for security
	}
	return nil
}
