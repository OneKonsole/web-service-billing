package helpers

import (
	"regexp"
	"unicode"
	"unicode/utf8"

	"github.com/go-playground/validator"
)

func isValidClusterName(fl validator.FieldLevel) bool {
	// Define the regular expression pattern
	clusterNamePattern := "^[a-z0-9][a-z0-9-]*[a-z0-9]$"

	// Compile the regular expression
	regex := regexp.MustCompile(clusterNamePattern)

	// Extract the field value
	clusterName := fl.Field().String()

	// Check if the clusterName matches the pattern
	return regex.MatchString(clusterName)
}

func startsWithAlphanum(fl validator.FieldLevel) bool {
	firstChar, _ := utf8.DecodeLastRuneInString(fl.Field().String())
	return unicode.IsLetter(firstChar) || unicode.IsDigit(firstChar)
}

func endWithAlphanum(fl validator.FieldLevel) bool {
	lastChar, _ := utf8.DecodeLastRuneInString(fl.Field().String())
	return unicode.IsLetter(lastChar) || unicode.IsDigit(lastChar)
}
func isUUID(fl validator.FieldLevel) bool {
	// Define the expected UUID format using a regular expression
	uuidPattern := regexp.MustCompile(`^[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}$`)

	// Check if the field matches the expected UUID format
	return uuidPattern.MatchString(fl.Field().String())
}
