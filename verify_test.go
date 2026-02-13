package shopifywebhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature_Valid(t *testing.T) {
	secret := "my-shopify-secret"
	body := []byte(`{"id":123,"email":"test@example.com"}`)
	signature := sign(secret, body)

	if err := VerifySignature(secret, body, signature); err != nil {
		t.Fatalf("expected valid signature, got error: %v", err)
	}
}

func TestVerifySignature_InvalidSignature(t *testing.T) {
	secret := "my-shopify-secret"
	body := []byte(`{"id":123}`)

	err := VerifySignature(secret, body, "bm90LWEtdmFsaWQtc2lnbmF0dXJl")
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got: %v", err)
	}
}

func TestVerifySignature_WrongSecret(t *testing.T) {
	body := []byte(`{"id":123}`)
	signature := sign("correct-secret", body)

	err := VerifySignature("wrong-secret", body, signature)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got: %v", err)
	}
}

func TestVerifySignature_TamperedBody(t *testing.T) {
	secret := "my-shopify-secret"
	originalBody := []byte(`{"id":123}`)
	signature := sign(secret, originalBody)

	tamperedBody := []byte(`{"id":456}`)
	err := VerifySignature(secret, tamperedBody, signature)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature for tampered body, got: %v", err)
	}
}

func TestVerifySignature_EmptyBody(t *testing.T) {
	secret := "my-shopify-secret"
	body := []byte("")
	signature := sign(secret, body)

	if err := VerifySignature(secret, body, signature); err != nil {
		t.Fatalf("expected valid signature for empty body, got: %v", err)
	}
}

func TestVerifyRequest_Valid(t *testing.T) {
	secret := "test-secret"
	body := `{"order_id":999}`
	signature := sign(secret, []byte(body))

	req := httptest.NewRequest("POST", "/webhooks", strings.NewReader(body))
	req.Header.Set("X-Shopify-Hmac-Sha256", signature)

	got, err := VerifyRequest(secret, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != body {
		t.Fatalf("expected body %q, got %q", body, string(got))
	}
}

func TestVerifyRequest_MissingHeader(t *testing.T) {
	req := httptest.NewRequest("POST", "/webhooks", strings.NewReader("{}"))
	// No X-Shopify-Hmac-Sha256 header.

	_, err := VerifyRequest("secret", req)
	if !errors.Is(err, ErrMissingSignature) {
		t.Fatalf("expected ErrMissingSignature, got: %v", err)
	}
}

func TestVerifyRequest_InvalidSignature(t *testing.T) {
	req := httptest.NewRequest("POST", "/webhooks", strings.NewReader(`{"id":1}`))
	req.Header.Set("X-Shopify-Hmac-Sha256", "aW52YWxpZA==")

	_, err := VerifyRequest("secret", req)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got: %v", err)
	}
}

func TestVerifyRequest_BodyConsumed(t *testing.T) {
	secret := "test-secret"
	body := `{"consumed":true}`
	signature := sign(secret, []byte(body))

	req := httptest.NewRequest("POST", "/webhooks", strings.NewReader(body))
	req.Header.Set("X-Shopify-Hmac-Sha256", signature)

	_, err := VerifyRequest(secret, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Body should be consumed after VerifyRequest.
	remaining, _ := io.ReadAll(req.Body)
	if len(remaining) != 0 {
		t.Fatalf("expected body to be consumed, got %d bytes", len(remaining))
	}
}
