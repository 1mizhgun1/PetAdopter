package utils

import (
	"fmt"
	"slices"
	"unicode/utf8"

	"pet_adopter/src/config"
)

var (
	extraUsernameChars = []int32{'_'}
	extraPasswordChars = []int32{'_', '!', '-'}
)

func isEnglishLetter(ch int32) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isDigit(ch int32) bool {
	return '0' <= ch && ch <= '9'
}

func ValidateUsername(username string, cfg config.ValidationConfig) error {
	length := utf8.RuneCountInString(username)

	if length < cfg.UsernameMinLength {
		return fmt.Errorf("username too short, minimum length: %d", cfg.UsernameMinLength)
	}

	if length > cfg.UsernameMaxLength {
		return fmt.Errorf("username too long, maximum length: %d", cfg.UsernameMaxLength)
	}

	for _, ch := range username {
		if !isEnglishLetter(ch) && !isDigit(ch) && !slices.Contains(extraUsernameChars, ch) {
			return fmt.Errorf("username can only have english letters, digits and extra characters: %v", extraUsernameChars)
		}
	}

	return nil
}

func ValidatePassword(password string, cfg config.ValidationConfig) error {
	length := utf8.RuneCountInString(password)

	if length < cfg.PasswordMinLength {
		return fmt.Errorf("password too short, minimum length: %d", cfg.PasswordMinLength)
	}

	if length > cfg.PasswordMaxLength {
		return fmt.Errorf("password too long, maximum length: %d", cfg.PasswordMaxLength)
	}

	for _, ch := range password {
		if !isEnglishLetter(ch) && !isDigit(ch) && !slices.Contains(extraPasswordChars, ch) {
			return fmt.Errorf("password can only have english letters, digits and extra characters: %v", extraPasswordChars)
		}
	}

	return nil
}
