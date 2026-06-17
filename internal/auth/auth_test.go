package auth

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestCheckJWTValid(t *testing.T) {
	userID := uuid.New()
	secretString := "TestCheckJWTValid"
	jwt, err := MakeJWT(userID, secretString, time.Second*2)
	if err != nil {
		t.Fatalf("error when creating jwt token: %v", err)
	}

	receivedID, err := ValidateJWT(jwt, secretString)
	if err != nil {
		t.Fatalf("error when validating jwt token: %v", err)
	}

	if receivedID != userID {
		t.Errorf("The IDs must be equal: %q != %q", userID, receivedID)
	}
}

func TestCheckJWTInvalid(t *testing.T) {
	userID := uuid.New()
	secret1 := "TestCheckJWTInvalid1"
	jwt, err := MakeJWT(userID, secret1, time.Second*2)
	if err != nil {
		t.Fatalf("error when creating jwt token: %v", err)
	}

	secret2 := "TectCheckJWTInvalid2"
	receivedID, err := ValidateJWT(jwt, secret2)
	if err == nil {
		t.Error("must have error when validating invalid jwt token")
	}

	if receivedID != uuid.Nil {
		t.Errorf("received ID must be uuid.Nil: %q", receivedID)
	}
}

func TestCheckExpiredJWT(t *testing.T) {
	userID := uuid.New()
	jwt, _ := MakeJWT(userID, "secret", -time.Second)
	receivedID, err := ValidateJWT(jwt, "secret")
	if receivedID != uuid.Nil {
		t.Errorf("received ID must be uuid.Nil")
	}

	if err == nil {
		t.Errorf("error must not be null")
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		header    http.Header
		wantToken string
		wantErr   bool
	}{
		{
			"valid header",
			http.Header{"Authorization": []string{"Bearer authentication_with_jwts"}},
			"authentication_with_jwts",
			false,
		},
		{
			"not have Bearer",
			http.Header{"Authorization": []string{" authentication_with_jwts"}},
			"",
			true,
		},
		{
			"not have token string",
			http.Header{"Authorization": []string{"Bearer "}},
			"",
			true,
		},
		{
			"missing authorization header",
			http.Header{},
			"",
			true,
		},
		{
			"header has extra spaces",
			http.Header{
				"Authorization": []string{"Bearer    authentication_with_jwts"},
			},
			"authentication_with_jwts",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			authHeader, err := GetBearerToken(test.header)
			getError := (err != nil)
			if getError != test.wantErr {
				t.Errorf("want: %v, actual: %v (check error)", test.wantErr, getError)
			}

			if authHeader != test.wantToken {
				t.Errorf("want: %q, actual: %q (check token)", test.wantToken, authHeader)
			}
		})
	}
}
