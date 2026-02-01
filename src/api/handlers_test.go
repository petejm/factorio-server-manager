package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OpenFactorioServerManager/factorio-server-manager/api"
)

func TestWriteResponse(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		wantCode int
	}{
		{
			name:     "string response",
			data:     "hello",
			wantCode: http.StatusOK,
		},
		{
			name:     "map response",
			data:     map[string]string{"key": "value"},
			wantCode: http.StatusOK,
		},
		{
			name:     "slice response",
			data:     []string{"a", "b", "c"},
			wantCode: http.StatusOK,
		},
		{
			name:     "struct response",
			data:     struct{ Name string }{"test"},
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			api.WriteResponse(rr, tt.data)

			// Response should be valid JSON
			var result interface{}
			if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
				t.Errorf("WriteResponse produced invalid JSON: %v", err)
			}
		})
	}
}

func TestReadRequestBody(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantErr    bool
		wantStatus int
	}{
		{
			name:       "valid body",
			body:       `{"key": "value"}`,
			wantErr:    false,
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty body",
			body:       "",
			wantErr:    false,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			body, _, err := api.ReadRequestBody(rr, req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadRequestBody() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && string(body) != tt.body {
				t.Errorf("ReadRequestBody() body = %q, want %q", string(body), tt.body)
			}
		})
	}
}

func TestReadRequestBodyNilBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", nil)
	req.Body = nil // Explicitly set to nil
	rr := httptest.NewRecorder()

	_, _, err := api.ReadRequestBody(rr, req)
	if err == nil {
		t.Error("ReadRequestBody() should return error for nil body")
	}

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for nil body, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestUnmarshallUserJson(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantUser   string
		wantErr    bool
		wantStatus int
	}{
		{
			name:     "valid user JSON",
			body:     `{"username": "testuser", "password": "testpass"}`,
			wantUser: "testuser",
			wantErr:  false,
		},
		{
			name:       "invalid JSON",
			body:       `{"username": testuser}`,
			wantErr:    true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "empty JSON object",
			body:     `{}`,
			wantUser: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			user, _, err := api.UnmarshallUserJson([]byte(tt.body), rr)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshallUserJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && user.Username != tt.wantUser {
				t.Errorf("UnmarshallUserJson() username = %q, want %q", user.Username, tt.wantUser)
			}
		})
	}
}

// Test helper to verify JSON response structure
func TestJSONResponseFileInputStructure(t *testing.T) {
	response := api.JSONResponseFileInput{
		Success:   true,
		Data:      "test data",
		Error:     "",
		ErrorKeys: []int{1, 2, 3},
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal JSONResponseFileInput: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify expected fields exist
	if _, ok := decoded["success"]; !ok {
		t.Error("Missing 'success' field in JSON response")
	}
	if _, ok := decoded["error"]; !ok {
		t.Error("Missing 'error' field in JSON response")
	}
	if _, ok := decoded["errorkeys"]; !ok {
		t.Error("Missing 'errorkeys' field in JSON response")
	}
}
