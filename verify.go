package shopifywebhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

// VerifySignature performs constant-time HMAC-SHA256 verification.
//
// Parameters:
//   - secret: the Shopify app's client secret
//   - body: the raw request body bytes
//   - signature: the base64-encoded value from X-Shopify-Hmac-Sha256
//
// Returns nil if valid, ErrInvalidSignature otherwise.
func VerifySignature(secret string, body []byte, signature string) error {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return ErrInvalidSignature
	}
	return nil
}

// VerifyRequest reads the request body, verifies the HMAC-SHA256 signature,
// and returns the raw body bytes.
//
// This is the key function that solves the common "parsed body breaks HMAC"
// problem: it reads the raw bytes first, verifies against those bytes, then
// returns them for JSON decoding.
//
// The request body is consumed. The returned bytes can be passed to
// json.Unmarshal or used with Event.Unmarshal.
func VerifyRequest(secret string, r *http.Request) ([]byte, error) {
	signature := r.Header.Get("X-Shopify-Hmac-Sha256")
	if signature == "" {
		return nil, ErrMissingSignature
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("shopifywebhook: reading request body: %w", err)
	}
	defer r.Body.Close()

	if err := VerifySignature(secret, body, signature); err != nil {
		return nil, err
	}
	return body, nil
}
