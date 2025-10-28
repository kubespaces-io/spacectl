package api

import (
	"encoding/json"
	"testing"
)

func TestRedactSensitiveJSON(t *testing.T) {
	input := []byte(`{
		"password": "secret",
		"refresh_token": "refresh",
		"nested": {
			"access_token": "token-value",
			"Authorization": "Bearer something",
			"safe": "value"
		},
		"list": [
			{"token": "abc123"},
			{"other": "value"}
		],
		"safe": "keep"
	}`)

	redacted := redactSensitiveJSON(input)
	var payload map[string]interface{}
	if err := json.Unmarshal(redacted, &payload); err != nil {
		t.Fatalf("redacted payload is not valid JSON: %v", err)
	}

	checkRedactedValue := func(m map[string]interface{}, key string) {
		if v, ok := m[key]; !ok || v != "***REDACTED***" {
			t.Fatalf("expected key %q to be redacted, got %v", key, v)
		}
	}

	checkRedactedValue(payload, "password")
	checkRedactedValue(payload, "refresh_token")

	nested, ok := payload["nested"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected nested field to be a map, got %T", payload["nested"])
	}
	checkRedactedValue(nested, "access_token")
	checkRedactedValue(nested, "Authorization")

	list, ok := payload["list"].([]interface{})
	if !ok {
		t.Fatalf("expected list field to be a slice, got %T", payload["list"])
	}
	entry, ok := list[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected list entry to be a map, got %T", list[0])
	}
	checkRedactedValue(entry, "token")

	if payload["safe"] != "keep" {
		t.Fatalf("expected non-sensitive key \"safe\" to retain its value, got %v", payload["safe"])
	}
	if nested["safe"] != "value" {
		t.Fatalf("expected nested non-sensitive key \"safe\" to retain its value, got %v", nested["safe"])
	}
}

func TestRedactSensitiveJSONHandlesInvalidJSON(t *testing.T) {
	input := []byte(`not-json`)
	redacted := redactSensitiveJSON(input)
	if string(redacted) != string(input) {
		t.Fatalf("expected invalid JSON input to be returned unchanged, got %q", redacted)
	}
}

func TestIsSensitiveKey(t *testing.T) {
	sensitive := []string{"password", "PASS", "Pwd", "access_token", "Refresh_Token", "token", "authorization"}
	for _, key := range sensitive {
		if !isSensitiveKey(key) {
			t.Fatalf("expected %q to be considered sensitive", key)
		}
	}

	nonSensitive := []string{"username", "email", "role"}
	for _, key := range nonSensitive {
		if isSensitiveKey(key) {
			t.Fatalf("expected %q to be considered non-sensitive", key)
		}
	}
}
