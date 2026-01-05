package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// LoadFixture loads a fixture from the fixtures directory
func LoadFixture(t *testing.T, filename string, key string) map[string]interface{} {
	// Get the project root (assuming tests/testutil is 2 levels deep)
	projectRoot := filepath.Join("..", "..")
	fixturePath := filepath.Join(projectRoot, "tests", "fixtures", filename)

	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read fixture file %s: %v", fixturePath, err)
	}

	var fixtures map[string]interface{}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatalf("Failed to parse fixture file %s: %v", fixturePath, err)
	}

	if key == "" {
		// Return entire fixture file
		return fixtures
	}

	// Return specific key
	if value, ok := fixtures[key]; ok {
		if valueMap, ok := value.(map[string]interface{}); ok {
			return valueMap
		}
		t.Fatalf("Fixture key %s is not a map", key)
	}

	t.Fatalf("Fixture key %s not found in %s", key, filename)
	return nil
}

// LoadFixtureRaw loads a fixture file as raw bytes
func LoadFixtureRaw(t *testing.T, filename string) []byte {
	projectRoot := filepath.Join("..", "..")
	fixturePath := filepath.Join(projectRoot, "tests", "fixtures", filename)

	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read fixture file %s: %v", fixturePath, err)
	}

	return data
}

