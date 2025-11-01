package setup

import (
	"testing"
	"time"
)

func TestParseDuration_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{
			name:     "seconds",
			input:    "30s",
			expected: 30 * time.Second,
		},
		{
			name:     "minutes",
			input:    "5m",
			expected: 5 * time.Minute,
		},
		{
			name:     "hours",
			input:    "2h",
			expected: 2 * time.Hour,
		},
		{
			name:     "milliseconds",
			input:    "500ms",
			expected: 500 * time.Millisecond,
		},
		{
			name:     "combined",
			input:    "1h30m",
			expected: 1*time.Hour + 30*time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err != nil {
				t.Fatalf("ParseDuration(%s) failed: %v", tt.input, err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParseDuration_Invalid(t *testing.T) {
	invalidInputs := []string{
		"",
		"invalid",
		"30",
		"abc123",
		"30x",
	}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			_, err := ParseDuration(input)
			if err == nil {
				t.Errorf("Expected error for invalid input %s, got nil", input)
			}
		})
	}
}

func TestParseRetryPolicy_Nil(t *testing.T) {
	result := ParseRetryPolicy(nil)
	if result != nil {
		t.Error("Expected nil for nil retries")
	}
}

func TestParseRetryPolicy_WithBackoff(t *testing.T) {
	retries := &Retries{
		Attempts: 3,
		Backoff:  "5s",
	}

	result := ParseRetryPolicy(retries)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Attempts != 3 {
		t.Errorf("Expected Attempts=3, got %d", result.Attempts)
	}

	if result.Backoff != 5*time.Second {
		t.Errorf("Expected Backoff=5s, got %v", result.Backoff)
	}
}

func TestParseRetryPolicy_WithoutBackoff(t *testing.T) {
	retries := &Retries{
		Attempts: 5,
		Backoff:  "",
	}

	result := ParseRetryPolicy(retries)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Attempts != 5 {
		t.Errorf("Expected Attempts=5, got %d", result.Attempts)
	}

	if result.Backoff != time.Second {
		t.Errorf("Expected default Backoff=1s, got %v", result.Backoff)
	}
}

func TestParseRetryPolicy_InvalidBackoff(t *testing.T) {
	retries := &Retries{
		Attempts: 3,
		Backoff:  "invalid",
	}

	result := ParseRetryPolicy(retries)
	if result == nil {
		t.Fatal("Expected non-nil result (should fallback to default)")
	}

	if result.Attempts != 3 {
		t.Errorf("Expected Attempts=3, got %d", result.Attempts)
	}

	if result.Backoff != time.Second {
		t.Errorf("Expected default Backoff=1s for invalid backoff, got %v", result.Backoff)
	}
}

func TestParseRetryPolicy_ComplexBackoff(t *testing.T) {
	retries := &Retries{
		Attempts: 10,
		Backoff:  "2m30s",
	}

	result := ParseRetryPolicy(retries)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	expectedBackoff := 2*time.Minute + 30*time.Second
	if result.Backoff != expectedBackoff {
		t.Errorf("Expected Backoff=%v, got %v", expectedBackoff, result.Backoff)
	}
}
