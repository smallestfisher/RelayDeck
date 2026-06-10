package auth

import "testing"

func TestHashPasswordAndVerify(t *testing.T) {
	hash, err := HashPassword("s3cret-password")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if !VerifyPassword(hash, "s3cret-password") {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Fatal("expected wrong password to fail verification")
	}
}
