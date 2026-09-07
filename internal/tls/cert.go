package tls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	ctls "crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func GenerateCert(domain, certsDir string, ca *x509.Certificate, caKey *ecdsa.PrivateKey) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: domain,
		},
		DNSNames:  []string{domain},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(825 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, ca, &key.PublicKey, caKey)
	if err != nil {
		return err
	}

	domainDir := filepath.Join(certsDir, domain)
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		return err
	}

	certFile, err := os.Create(filepath.Join(domainDir, "cert.pem"))
	if err != nil {
		return err
	}
	defer certFile.Close()
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}

	keyFile, err := os.OpenFile(filepath.Join(domainDir, "key.pem"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyFile.Close()
	pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return nil
}

func LoadCert(certsDir, domain string) (ctls.Certificate, error) {
	certPath := filepath.Join(certsDir, domain, "cert.pem")
	keyPath := filepath.Join(certsDir, domain, "key.pem")
	return ctls.LoadX509KeyPair(certPath, keyPath)
}
