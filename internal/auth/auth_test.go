package auth

import (
	"strings"
	"testing"
)

func TestCheckPasswordHashValid(t *testing.T) {
	inputs := []string{"hello", "world", "hello world"}
	for _, input := range inputs {
		hashed, err := HashPassword(input)
		if err != nil {
			t.Fatalf("error when hashing password: %v", err)
		}

		isValid, err := CheckPasswordHash(input, hashed)
		if err != nil {
			t.Fatalf("error when checking password: %v", err)
		}

		if !isValid {
			t.Errorf("HashPassword(%q) has some error, its output: %q", input, hashed)
		}
	}
}

func TestCheckPasswordHashInvalid(t *testing.T) {
	inputs := []string{"hello", "WORLD", "HELLO world"}
	for _, input := range inputs {
		hashed, err := HashPassword(input)
		if err != nil {
			t.Fatalf("error when hashing password: %v", err)
		}

		wrongPassword := strings.ToLower(input) + input + strings.ToUpper(input)
		isValid, err := CheckPasswordHash(wrongPassword, hashed)
		if err != nil {
			t.Fatalf("error when checking password: %v", err)
		}

		if isValid {
			t.Errorf("HashPassword(%q) or CheckPasswordHash(%q) has some error, %q != %q", input, input, input, wrongPassword)
		}
	}
}
