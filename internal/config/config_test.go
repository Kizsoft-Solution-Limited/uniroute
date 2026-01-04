package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original env
	originalPort := os.Getenv("PORT")
	originalOllama := os.Getenv("OLLAMA_BASE_URL")

	// Clean up after test
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
		if originalOllama != "" {
			os.Setenv("OLLAMA_BASE_URL", originalOllama)
		} else {
			os.Unsetenv("OLLAMA_BASE_URL")
		}
	}()

	// Test default values
	os.Unsetenv("PORT")
	os.Unsetenv("OLLAMA_BASE_URL")

	cfg := Load()
	if cfg.Port != "8084" {
		t.Errorf("Expected default port '8084', got '%s'", cfg.Port)
	}
	if cfg.OllamaBaseURL != "http://localhost:11434" {
		t.Errorf("Expected default Ollama URL 'http://localhost:11434', got '%s'", cfg.OllamaBaseURL)
	}
}

func TestLoad_WithEnvVars(t *testing.T) {
	// Save original env
	originalPort := os.Getenv("PORT")
	originalOllama := os.Getenv("OLLAMA_BASE_URL")

	// Clean up after test
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
		if originalOllama != "" {
			os.Setenv("OLLAMA_BASE_URL", originalOllama)
		} else {
			os.Unsetenv("OLLAMA_BASE_URL")
		}
	}()

	// Test with environment variables
	os.Setenv("PORT", "9999")
	os.Setenv("OLLAMA_BASE_URL", "http://custom:11434")

	cfg := Load()
	if cfg.Port != "9999" {
		t.Errorf("Expected port '9999', got '%s'", cfg.Port)
	}
	if cfg.OllamaBaseURL != "http://custom:11434" {
		t.Errorf("Expected Ollama URL 'http://custom:11434', got '%s'", cfg.OllamaBaseURL)
	}
}

func TestGetEnv(t *testing.T) {
	// Save original env
	original := os.Getenv("TEST_VAR")

	// Clean up after test
	defer func() {
		if original != "" {
			os.Setenv("TEST_VAR", original)
		} else {
			os.Unsetenv("TEST_VAR")
		}
	}()

	// Test with default value
	os.Unsetenv("TEST_VAR")
	got := getEnv("TEST_VAR", "default")
	if got != "default" {
		t.Errorf("Expected 'default', got '%s'", got)
	}

	// Test with set value
	os.Setenv("TEST_VAR", "custom")
	got = getEnv("TEST_VAR", "default")
	if got != "custom" {
		t.Errorf("Expected 'custom', got '%s'", got)
	}
}

