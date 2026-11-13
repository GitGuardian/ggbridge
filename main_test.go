package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
	return path
}

func TestCaBundleContents(t *testing.T) {
	dir := t.TempDir()
	system := writeFile(t, dir, "ca-certificates.crt", "SYSTEM\n")
	private := writeFile(t, dir, "private.crt", "PRIVATE\n")
	mtls := writeFile(t, dir, "mtls.crt", "MTLS\n")
	empty := writeFile(t, dir, "empty.crt", "")
	missing := filepath.Join(dir, "missing.crt")

	tests := []struct {
		name     string
		extra    []string
		expected string
	}{
		{"no extra certificates", nil, "SYSTEM\n"},
		{"private CA bundle", []string{private}, "SYSTEM\nPRIVATE\n"},
		{"private and mTLS CA", []string{private, mtls}, "SYSTEM\nPRIVATE\nMTLS\n"},
		{"missing certificates are skipped", []string{missing, mtls}, "SYSTEM\nMTLS\n"},
		{"empty certificates are skipped", []string{empty, mtls}, "SYSTEM\nMTLS\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle, err := caBundleContents(system, tt.extra)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(bundle) != tt.expected {
				t.Errorf("got %q, want %q", bundle, tt.expected)
			}
		})
	}
}

func TestCaBundleContentsAddsMissingNewline(t *testing.T) {
	dir := t.TempDir()
	system := writeFile(t, dir, "ca-certificates.crt", "SYSTEM")
	private := writeFile(t, dir, "private.crt", "PRIVATE\n")

	bundle, err := caBundleContents(system, []string{private})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(bundle) != "SYSTEM\nPRIVATE\n" {
		t.Errorf("got %q, want %q", bundle, "SYSTEM\nPRIVATE\n")
	}
}

func TestCaBundleContentsMissingSystemBundle(t *testing.T) {
	if _, err := caBundleContents(filepath.Join(t.TempDir(), "missing.crt"), nil); err == nil {
		t.Fatal("expected an error for a missing system CA bundle")
	}
}

func TestBuildCaBundle(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "certs", "ca-bundle.crt")

	t.Setenv("SSL_CERT_FILE", writeFile(t, dir, "ca-certificates.crt", "SYSTEM\n"))
	t.Setenv("GGBRIDGE_SSL_CERT_DIR", filepath.Dir(target))
	t.Setenv("GGBRIDGE_SSL_CERT_FILE", target)
	t.Setenv("GGBRIDGE_SSL_PRIVATE_CERT_FILE", writeFile(t, dir, "private.crt", "PRIVATE\n"))
	t.Setenv("GGBRIDGE_MTLS_CA_FILE", filepath.Join(dir, "missing.crt"))

	if err := buildCaBundle(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bundle, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read the generated bundle: %v", err)
	}
	if string(bundle) != "SYSTEM\nPRIVATE\n" {
		t.Errorf("got %q, want %q", bundle, "SYSTEM\nPRIVATE\n")
	}
}

func TestBuildCaBundleIgnoresSelfReferencingSystemBundle(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "ca-bundle.crt")
	writeFile(t, dir, "ca-bundle.crt", "STALE\n")

	t.Setenv("SSL_CERT_FILE", target)
	t.Setenv("GGBRIDGE_SSL_CERT_DIR", dir)
	t.Setenv("GGBRIDGE_SSL_CERT_FILE", target)
	t.Setenv("GGBRIDGE_SSL_PRIVATE_CERT_FILE", filepath.Join(dir, "missing.crt"))
	t.Setenv("GGBRIDGE_MTLS_CA_FILE", filepath.Join(dir, "missing.crt"))

	if _, err := os.Stat(DefaultSystemCaBundleFile); err != nil {
		t.Skipf("no system CA bundle at %s", DefaultSystemCaBundleFile)
	}

	if err := buildCaBundle(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bundle, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read the generated bundle: %v", err)
	}
	if bytes.Contains(bundle, []byte("STALE")) {
		t.Error("the generated bundle was fed back into itself")
	}
}
