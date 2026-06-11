package secretbox

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	box, err := New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("new box: %v", err)
	}

	ciphertext, err := box.Encrypt("sk-live-secret")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if ciphertext == "" {
		t.Fatal("expected ciphertext")
	}
	if ciphertext == "sk-live-secret" {
		t.Fatal("ciphertext must not equal plaintext")
	}

	plaintext, err := box.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if plaintext != "sk-live-secret" {
		t.Fatalf("expected plaintext round trip, got %q", plaintext)
	}
}

func TestEncryptDecryptEmptySecret(t *testing.T) {
	box, err := New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("new box: %v", err)
	}

	ciphertext, err := box.Encrypt("")
	if err != nil {
		t.Fatalf("encrypt empty secret: %v", err)
	}

	plaintext, err := box.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt empty secret: %v", err)
	}
	if plaintext != "" {
		t.Fatalf("expected empty plaintext, got %q", plaintext)
	}
}

func TestNewRejectsInvalidKey(t *testing.T) {
	if _, err := New([]byte("short")); err == nil {
		t.Fatal("expected invalid key error")
	}
}

func TestDecryptRejectsWrongKey(t *testing.T) {
	box, err := New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("new box: %v", err)
	}
	otherBox, err := New([]byte("abcdef0123456789abcdef0123456789"))
	if err != nil {
		t.Fatalf("new other box: %v", err)
	}

	ciphertext, err := box.Encrypt("sk-live-secret")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if _, err := otherBox.Decrypt(ciphertext); err == nil {
		t.Fatal("expected wrong-key decrypt error")
	}
}
