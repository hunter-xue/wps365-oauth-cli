// Package sign implements the KSO-1 request signing algorithm for WPS365 Open APIs.
//
// Reference: https://365.kdocs.cn/3rd/open/documents/app-integration-dev/wps365/server/api-description/signature-description
//
// Signature format:
//
//	X-Kso-Authorization: KSO-1 {accessKey}:{signature}
//
// Where signature = HMAC-SHA256(secretKey,
//
//	"KSO-1" + Method + RequestURI + ContentType + KsoDate + sha256hex(RequestBody))
//
// RequestBody sha256hex is empty string "" when body is empty.
package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// KsoSign computes KSO-1 signatures for HTTP requests.
type KsoSign struct {
	accessKey string // APP_ID
	secretKey string // APP SECRET
}

// Out holds the computed header values to set on the request.
type Out struct {
	Date          string // X-Kso-Date
	Authorization string // X-Kso-Authorization
}

// New creates a KsoSign. Both accessKey and secretKey are required.
func New(accessKey, secretKey string) (*KsoSign, error) {
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("sign.New: accessKey and secretKey must not be empty")
	}
	return &KsoSign{accessKey: accessKey, secretKey: secretKey}, nil
}

// Sign computes the KSO-1 signature for an outgoing HTTP request.
// body is the raw request body (may be nil or empty for GET/DELETE).
// The returned Out contains the values for X-Kso-Date and X-Kso-Authorization.
func (k *KsoSign) Sign(req *http.Request, body []byte) (*Out, error) {
	contentType := req.Header.Get("Content-Type")
	ksoDate := time.Now().UTC().Format(time.RFC1123)
	sig := k.computeSignature(req.Method, req.URL.RequestURI(), contentType, ksoDate, body)
	return &Out{
		Date:          ksoDate,
		Authorization: fmt.Sprintf("KSO-1 %s:%s", k.accessKey, sig),
	}, nil
}

// Apply is a convenience method that calls Sign and sets the headers directly on req.
func (k *KsoSign) Apply(req *http.Request, body []byte) error {
	out, err := k.Sign(req, body)
	if err != nil {
		return err
	}
	req.Header.Set("X-Kso-Date", out.Date)
	req.Header.Set("X-Kso-Authorization", out.Authorization)
	return nil
}

// computeSignature is the core HMAC-SHA256 computation.
//
// Signed string: "KSO-1" + Method + RequestURI + ContentType + KsoDate + sha256hex(body)
func (k *KsoSign) computeSignature(method, requestURI, contentType, ksoDate string, body []byte) string {
	sha256Hex := ""
	if len(body) > 0 {
		h := sha256.New()
		h.Write(body)
		sha256Hex = hex.EncodeToString(h.Sum(nil))
	}

	payload := "KSO-1" + method + requestURI + contentType + ksoDate + sha256Hex

	mac := hmac.New(sha256.New, []byte(k.secretKey))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
