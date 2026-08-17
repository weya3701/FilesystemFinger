package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ccxn/filesystemfinger/fingerprint"
)

func TestScanAndVerify(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	var scanOutput bytes.Buffer
	var stderr bytes.Buffer
	if err := run([]string{"scan", root}, &scanOutput, &stderr); err != nil {
		t.Fatalf("scan failed: %v (%s)", err, stderr.String())
	}
	var manifest fingerprint.Manifest
	if err := json.Unmarshal(scanOutput.Bytes(), &manifest); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(manifestPath, scanOutput.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}

	var verifyOutput bytes.Buffer
	if err := run([]string{"verify", "--manifest", manifestPath, root}, &verifyOutput, &stderr); err != nil {
		t.Fatalf("verify failed: %v (%s)", err, stderr.String())
	}
	var result fingerprint.Verification
	if err := json.Unmarshal(verifyOutput.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Match {
		t.Fatal("unchanged directory did not verify")
	}
}

func TestScanHashOnly(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	want, err := fingerprint.Hash(root, fingerprint.Options{UseDefaults: true})
	if err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := run([]string{"scan", "--hash-only", root}, &stdout, &stderr); err != nil {
		t.Fatalf("scan --hash-only failed: %v (%s)", err, stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); got != want {
		t.Fatalf("hash-only output = %q, want %q", got, want)
	}
}

func TestVerifyDetectsRename(t *testing.T) {
	root := t.TempDir()
	original := filepath.Join(root, "before.txt")
	if err := os.WriteFile(original, []byte("same"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest, err := fingerprint.Scan(root, fingerprint.Options{UseDefaults: true})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(manifestPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(original, filepath.Join(root, "after.txt")); err != nil {
		t.Fatal(err)
	}

	err = run([]string{"verify", "--manifest", manifestPath, root}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("rename was not detected")
	}
}
