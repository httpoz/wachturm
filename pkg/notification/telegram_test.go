package notification

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestTelegramNotifier_Send(t *testing.T) {
	// Create a test server that simulates the Telegram API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		// Verify that the body contains the expected message
		bodyStr := string(body)
		if bodyStr == "" {
			t.Error("Request body is empty")
		}

		// Check that it contains the chat_id and text fields
		if !contains(bodyStr, "chat_id") || !contains(bodyStr, "text") {
			t.Errorf("Request body missing required fields: %s", bodyStr)
		}

		// Write successful response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer server.Close()

	// Create a test notifier pointing to our test server
	notifier := NewTelegramNotifier("test_token", "test_chat_id")

	// Override the httpClient to use our test server URL
	notifier.httpClient = &http.Client{}
	notifier.baseURL = server.URL
	// Send a test message
	err := notifier.Send("Test message")
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
}

func TestTelegramNotifier_SendSummary(t *testing.T) {
	// Create a temporary file with test content
	tempFile, err := os.CreateTemp("", "summary_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testContent := "Test summary content"
	if _, err := tempFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Create a test server that simulates the Telegram API
	var receivedMessage string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedMessage = string(body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	// Create test notifier
	notifier := NewTelegramNotifier("test_token", "test_chat_id")
	notifier.httpClient = &http.Client{}
	notifier.baseURL = server.URL

	// Send the summary
	err = notifier.SendSummary(tempFile.Name())
	if err != nil {
		t.Fatalf("SendSummary returned error: %v", err)
	}

	// Verify the sent message contains our test content
	if !contains(receivedMessage, testContent) {
		t.Errorf("Sent message doesn't contain the summary content. Message: %s", receivedMessage)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr || len(s) > len(substr) && s[len(s)-len(substr):] == substr || len(s) > len(substr) && s[0:len(substr)] != substr && s[len(s)-len(substr):] != substr && len(s) > 0 && contains(s[1:], substr)
}
