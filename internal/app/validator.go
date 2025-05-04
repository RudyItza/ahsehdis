package app

import (
	"regexp"
	"strings"
)
// emailRegex is a compiled regular expression for validating email format.
var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)
// Validator is a structure used for collecting validation errors.
type Validator struct {
	Errors map[string]string // A map of validation errors with keys and error messages.
}
// NewValidator creates and returns a new Validator instance.
func NewValidator() *Validator {
	// Initialize a new Validator instance with an empty Errors map.
	return &Validator{Errors: make(map[string]string)}
}
// Valid checks if the Validator has any errors.
func (v *Validator) Valid() bool {
	// Returns true if there are no errors in the Errors map.
	return len(v.Errors) == 0
}
// AddError adds a validation error to the Validator's Errors map if the error key doesn't already exist.
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}
// Check adds an error to the Validator if the condition (ok) is false.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}
// NotBlank checks if a given string is not blank (i.e., not just whitespace).
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}
// MatchesPattern checks if a given string matches the specified regular expression.
func MatchesPattern(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
// ValidateEmail checks if an email address is valid based on the emailRegex pattern.
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}
