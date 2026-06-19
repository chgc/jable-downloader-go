package config

import (
	"testing"
)

func TestConstants(t *testing.T) {
	t.Run("UserAgent_should_not_be_empty", func(t *testing.T) {
		if UserAgent == "" {
			t.Error("UserAgent should not be empty")
		}
	})

	t.Run("UserAgent_should_contain_Chrome", func(t *testing.T) {
		if UserAgent != "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.97 Safari/537.36" {
			t.Error("UserAgent should match expected value")
		}
	})

	t.Run("MaxWorkers_should_be_positive", func(t *testing.T) {
		if MaxWorkers <= 0 {
			t.Errorf("MaxWorkers should be positive, got %d", MaxWorkers)
		}
	})

	t.Run("MaxWorkers_should_be_8", func(t *testing.T) {
		if MaxWorkers != 8 {
			t.Errorf("MaxWorkers should be 8, got %d", MaxWorkers)
		}
	})

	t.Run("Headers_should_contain_UserAgent", func(t *testing.T) {
		if val, ok := Headers["User-Agent"]; !ok {
			t.Error("Headers should contain User-Agent key")
		} else if val != UserAgent {
			t.Errorf("Headers['User-Agent'] should be %q, got %q", UserAgent, val)
		}
	})

	t.Run("Headers_should_have_length_1", func(t *testing.T) {
		if len(Headers) != 1 {
			t.Errorf("Headers should have 1 entry, got %d", len(Headers))
		}
	})
}
