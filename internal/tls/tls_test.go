package tls

import (
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCACreatesFiles(t *testing.T) {
	dir := t.TempDir()
	if err := InitCA(dir); err != nil {
		t.Fatalf("InitCA() error: %v", err)
	}

	certPath := filepath.Join(dir, "ca-cert.pem")
	keyPath := filepath.Join(dir, "ca-key.pem")

	if _, err := os.Stat(certPath); err != nil {
		t.Errorf("CA cert not created: %v", err)
	}
	if _, err := os.Stat(keyPath); err != nil {
		t.Errorf("CA key not created: %v", err)
	}
}

func TestLoadCA(t *testing.T) {
	dir := t.TempDir()
	InitCA(dir)

	ca, key, err := LoadCA(dir)
	if err != nil {
		t.Fatalf("LoadCA() error: %v", err)
	}
	if ca == nil {
		t.Fatal("CA cert is nil")
	}
	if key == nil {
		t.Fatal("CA key is nil")
	}
	if !ca.IsCA {
		t.Error("cert is not a CA")
	}
}

func TestCAExists(t *testing.T) {
	dir := t.TempDir()
	if CAExists(dir) {
		t.Error("CAExists should be false before InitCA")
	}
	InitCA(dir)
	if !CAExists(dir) {
		t.Error("CAExists should be true after InitCA")
	}
}

func TestGenerateAndLoadCert(t *testing.T) {
	dir := t.TempDir()
	certsDir := filepath.Join(dir, "certs")
	InitCA(dir)

	ca, caKey, _ := LoadCA(dir)

	if err := GenerateCert("myapp.test", certsDir, ca, caKey); err != nil {
		t.Fatalf("GenerateCert() error: %v", err)
	}

	certFile := filepath.Join(certsDir, "myapp.test", "cert.pem")
	keyFile := filepath.Join(certsDir, "myapp.test", "key.pem")

	if _, err := os.Stat(certFile); err != nil {
		t.Errorf("domain cert not created: %v", err)
	}
	if _, err := os.Stat(keyFile); err != nil {
		t.Errorf("domain key not created: %v", err)
	}

	tlsCert, err := LoadCert(certsDir, "myapp.test")
	if err != nil {
		t.Fatalf("LoadCert() error: %v", err)
	}

	leaf, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}

	pool := x509.NewCertPool()
	pool.AddCert(ca)
	if _, err := leaf.Verify(x509.VerifyOptions{Roots: pool}); err != nil {
		t.Errorf("cert not signed by CA: %v", err)
	}
}

func TestGenerateCertIncludesDNSName(t *testing.T) {
	dir := t.TempDir()
	certsDir := filepath.Join(dir, "certs")
	InitCA(dir)
	ca, caKey, _ := LoadCA(dir)
	GenerateCert("cayacart.test", certsDir, ca, caKey)

	tlsCert, _ := LoadCert(certsDir, "cayacart.test")
	leaf, _ := x509.ParseCertificate(tlsCert.Certificate[0])

	found := false
	for _, name := range leaf.DNSNames {
		if name == "cayacart.test" {
			found = true
		}
	}
	if !found {
		t.Errorf("cert DNSNames %v should contain cayacart.test", leaf.DNSNames)
	}
}
