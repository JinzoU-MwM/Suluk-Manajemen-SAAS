package service

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/jamaah-in/v2/internal/auth/model"
)

func TestEvaluateGoogleLink(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-horse"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate hash: %v", err)
	}
	existing := &model.User{PasswordHash: string(hash)}

	t.Run("no password supplied", func(t *testing.T) {
		if err := evaluateGoogleLink(existing, ""); err != ErrGoogleLinkPasswordRequired {
			t.Errorf("got %v, want ErrGoogleLinkPasswordRequired", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		if err := evaluateGoogleLink(existing, "wrong-password"); err != ErrGoogleLinkPasswordIncorrect {
			t.Errorf("got %v, want ErrGoogleLinkPasswordIncorrect", err)
		}
	})

	t.Run("correct password", func(t *testing.T) {
		if err := evaluateGoogleLink(existing, "correct-horse"); err != nil {
			t.Errorf("got %v, want nil", err)
		}
	})
}
