package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

func parseJSONObject(raw string) (map[string]any, error) {
	if raw == "" {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("parse JSON object: %w", err)
	}
	return out, nil
}

func parseJSONArrayObjects(raw string) ([]map[string]any, error) {
	if raw == "" {
		return nil, nil
	}
	var out []map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("parse JSON array: %w", err)
	}
	return out, nil
}

func parseJSONObjectFile(path string) (map[string]any, error) {
	if path == "" {
		return map[string]any{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseJSONObject(string(data))
}
