package rules

import (
	"strings"
	"unicode"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/YagoSchramm/GoDepot/domain/entity/derr"
)

func ValidateLogin(credentials entity.UserCredentials) error {
	err := validateEmail(credentials.Email)
	if err != nil {
		return err
	}
	if strings.TrimSpace(credentials.Password) == "" {
		return derr.PasswordRequired
	}

	return nil
}
func ValidateRegister(user entity.User) error {
	// TODO: Add later the other user fields validation
	err := validateEmail(user.Email)
	if err != nil {
		return err
	}
	err = validatePassword(user.Password)
	if err != nil {
		return err
	}
	return nil
}
func validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return derr.EmailRequired
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" {
		return derr.InvalidEmail
	}

	domain := parts[1]
	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return derr.InvalidEmail
	}

	return nil
}

func validatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return derr.PasswordRequired
	}

	if len(password) < 8 {
		return derr.WeakPassword
	}

	hasLetter := false
	hasDigit := false
	for _, ch := range password {
		if unicode.IsLetter(ch) {
			hasLetter = true
		}
		if unicode.IsDigit(ch) {
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return derr.WeakPassword
	}

	return nil
}
