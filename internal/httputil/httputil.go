package httputil

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/DylanBlakemore/initiat-cli/internal/encoding"
	"github.com/DylanBlakemore/initiat-cli/internal/storage"
	"github.com/DylanBlakemore/initiat-cli/internal/types"
)

const (
	UserAgent = "initiat-cli/1.0"
)

func ParseAPIResponse(body []byte, target interface{}) error {
	var apiResp types.APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			return fmt.Errorf("API error: %s", apiResp.Errors[0])
		}
		return fmt.Errorf("API error: %s", apiResp.Message)
	}

	if target != nil && len(apiResp.Data) > 0 {
		if err := json.Unmarshal(apiResp.Data, target); err != nil {
			return fmt.Errorf("failed to parse response data: %w", err)
		}
	}

	return nil
}

func ParseValidationErrorResponse(body []byte) error {
	var validationResp types.ValidationErrorResponse
	if err := json.Unmarshal(body, &validationResp); err != nil {
		var apiResp types.APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return fmt.Errorf("failed to parse error response: %w", err)
		}
		if len(apiResp.Errors) > 0 {
			return fmt.Errorf("validation error: %s", apiResp.Errors[0])
		}
		return fmt.Errorf("validation error: %s", apiResp.Message)
	}

	if validationResp.Success {
		return nil
	}
	if len(validationResp.Errors) > 0 {
		var errorMessages []string
		for field, messages := range validationResp.Errors {
			for _, msg := range messages {
				errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, msg))
			}
		}
		if len(errorMessages) > 0 {
			return fmt.Errorf("validation failed: %s", errorMessages[0])
		}
	}

	return fmt.Errorf("validation error: %s", validationResp.Message)
}

func SetCommonHeaders(req *http.Request, contentType string) {
	req.Header.Set("User-Agent", UserAgent)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
}

func SignRequest(req *http.Request, _ []byte) error {
	store := storage.New()

	deviceID, err := store.GetDeviceID()
	if err != nil {
		return fmt.Errorf("failed to get device ID: %w", err)
	}

	signingKey, err := store.GetSigningPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to get signing key: %w", err)
	}

	timestamp := time.Now().Unix()

	message := fmt.Sprintf("%s\n%s\n%d",
		req.Method,
		req.URL.Path+req.URL.RawQuery,
		timestamp)

	signature := ed25519.Sign(signingKey, []byte(message))

	signatureEncoded, err := encoding.EncodeEd25519Signature(signature)
	if err != nil {
		return fmt.Errorf("failed to encode signature: %w", err)
	}

	req.Header.Set("Authorization", "Device "+deviceID)
	req.Header.Set("X-Signature", signatureEncoded)
	req.Header.Set("X-Timestamp", strconv.FormatInt(timestamp, 10))

	return nil
}

func DoSignedRequest(client *http.Client, method, url string, body []byte) (int, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	contentType := ""
	if body != nil {
		contentType = "application/json"
	}
	SetCommonHeaders(req, contentType)

	if err := SignRequest(req, body); err != nil {
		return 0, nil, fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, respBody, nil
}

func DoUnsignedRequest(client *http.Client, method, url string, body []byte) (int, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	contentType := ""
	if body != nil {
		contentType = "application/json"
	}
	SetCommonHeaders(req, contentType)

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, respBody, nil
}

func HandleStandardResponse(statusCode int, body []byte, target interface{}) error {
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return ParseValidationErrorResponse(body)
	}

	return ParseAPIResponse(body, target)
}

func HandleGetResponse(statusCode int, body []byte, target interface{}) error {
	if statusCode != http.StatusOK {
		return ParseValidationErrorResponse(body)
	}

	return ParseAPIResponse(body, target)
}

func HandleDeleteResponse(statusCode int, body []byte) error {
	if statusCode != http.StatusOK && statusCode != http.StatusNoContent {
		return ParseValidationErrorResponse(body)
	}

	return ParseAPIResponse(body, nil)
}
