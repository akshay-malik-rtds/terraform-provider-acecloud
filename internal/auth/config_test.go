package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadCLIConfig_Valid(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{
		"api_base_url": "https://api.acecloud.ai",
		"token": "eyJhbGciOiJSUzI1NiJ9.test-token",
		"region": "mumbai",
		"project_id": "proj-abc-123"
	}`), 0600)

	result, err := readCLIConfig(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Token != "eyJhbGciOiJSUzI1NiJ9.test-token" {
		t.Errorf("expected token eyJhbGciOiJSUzI1NiJ9.test-token, got %s", result.Token)
	}
	if result.Region != "mumbai" {
		t.Errorf("expected region mumbai, got %s", result.Region)
	}
	if result.ProjectID != "proj-abc-123" {
		t.Errorf("expected project_id proj-abc-123, got %s", result.ProjectID)
	}
	if result.Method != "cli_config" {
		t.Errorf("expected method cli_config, got %s", result.Method)
	}
}

func TestReadCLIConfig_MissingFile(t *testing.T) {
	_, err := readCLIConfig("/tmp/nonexistent-ace-config-12345/config.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if got := err.Error(); got == "" {
		t.Error("expected non-empty error message")
	}
}

func TestReadCLIConfig_EmptyToken(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{
		"api_base_url": "https://api.acecloud.ai",
		"token": "",
		"region": "mumbai",
		"project_id": "proj-1"
	}`), 0600)

	_, err := readCLIConfig(cfgPath)
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestReadCLIConfig_CorruptJSON(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{not json`), 0600)

	_, err := readCLIConfig(cfgPath)
	if err == nil {
		t.Fatal("expected error for corrupt JSON")
	}
}

func TestReadCLIConfig_DefaultPath(t *testing.T) {
	// With empty path, it should try ~/.ace/config.json.
	// This may or may not exist, but we just verify it doesn't panic.
	_, _ = readCLIConfig("")
}

func TestReadCLIConfigAPIURL_Valid(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{
		"api_base_url": "https://api.acecloud.ai",
		"token": "tok"
	}`), 0600)

	url := ReadCLIConfigAPIURL(cfgPath)
	if url != "https://api.acecloud.ai" {
		t.Errorf("expected https://api.acecloud.ai, got %s", url)
	}
}

func TestReadCLIConfigAPIURL_Missing(t *testing.T) {
	url := ReadCLIConfigAPIURL("/tmp/nonexistent-12345/config.json")
	if url != "" {
		t.Errorf("expected empty string, got %s", url)
	}
}
