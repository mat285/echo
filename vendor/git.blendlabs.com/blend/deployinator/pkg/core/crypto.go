package core

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"

	"github.com/blend/go-sdk/crypto"
	exception "github.com/blend/go-sdk/exception"
)

// Crypto is a namespace for crypto related functions.
var Crypto = cryptoUtil{}

type cryptoUtil struct{}

// GenerateBase64Secret generates a secret numBytes long and base64 encodes it
func (cr cryptoUtil) GenerateBase64Secret(numBytes int) (string, error) {
	bytes, err := crypto.CreateKey(numBytes)
	if err != nil {
		return "", exception.New(err)
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// PemDecode decodes the pem data into a block or errors
func PemDecode(pemData []byte) (*pem.Block, error) {
	b, _ := pem.Decode(pemData)
	if b == nil {
		return nil, exception.New(fmt.Sprintf("No certificate found"))
	}
	return b, nil
}

// ParseCertificate parses the pem encoded data into an x509 certificate
func ParseCertificate(pemData []byte) (*x509.Certificate, error) {
	b, err := PemDecode(pemData)
	if err != nil {
		return nil, exception.New(err)
	}
	cert, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		return nil, exception.New(err)
	}
	return cert, nil
}

// ParsePrivateKey parses the pem encoded private key
func ParsePrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	b, err := PemDecode(pemData)
	if err != nil {
		return nil, exception.New(err)
	}
	priv, err := x509.ParsePKCS1PrivateKey(b.Bytes)
	if err != nil {
		return TryParsePKCS8PrivateKey(b)
	}
	return priv, nil
}

// TryParsePKCS8PrivateKey tries to parse the rsa private key from the data
func TryParsePKCS8PrivateKey(b *pem.Block) (*rsa.PrivateKey, error) {
	key, err := x509.ParsePKCS8PrivateKey(b.Bytes)
	if err != nil {
		return nil, exception.New(err)
	}
	priv, ok := key.(*rsa.PrivateKey)
	if !ok || priv == nil {
		return nil, exception.New("Not an RSA private key")
	}
	return priv, nil
}

// CertificateModulus returns the modulus of the certificate
func CertificateModulus(cert *x509.Certificate) (*big.Int, error) {
	pub, err := x509.ParsePKIXPublicKey(cert.RawSubjectPublicKeyInfo)
	if err != nil {
		return nil, exception.New(err)
	}
	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub.N, nil
	}
	return nil, exception.New(fmt.Sprintf("Unable to parse rsa public key"))
}

// VerifyTLSCertPair verifies that the tls cert pair match each other
func VerifyTLSCertPair(cert []byte, key []byte) error {
	certificate, err := ParseCertificate(cert)
	if err != nil {
		return exception.New(err)
	}
	priv, err := ParsePrivateKey(key)
	if err != nil {
		return exception.New(err)
	}
	mod, err := CertificateModulus(certificate)
	if err != nil {
		return exception.New(err)
	}
	if mod.Cmp(priv.N) != 0 {
		return exception.New(fmt.Sprintf("Cert and key do not match"))
	}
	return nil
}

// VerifyCertDomain verifies that the cert is valid for the domain at the current time
func VerifyCertDomain(cert []byte, dns string) (bool, []string, error) {
	certificate, err := ParseCertificate(cert)
	if err != nil {
		return false, nil, exception.New(err)
	}
	names := CertificateDomainNames(certificate)
	for _, name := range names {
		if DoDomainsMatch(name, dns) {
			return true, names, nil
		}
	}
	return false, names, nil
}

// DoDomainsMatch returns if the cert domain name matches the name including wildcards
func DoDomainsMatch(certDNS, name string) bool {
	certDNS = strings.TrimSpace(strings.ToLower(certDNS))
	name = strings.TrimSpace(strings.ToLower(name))
	if strings.HasPrefix(certDNS, "*.") { // trim wildcards
		return strings.HasSuffix(name, strings.TrimPrefix(certDNS, "*."))
	}
	return name == certDNS
}

// CertificateDomainNames returns the valid domain names for this certificate
func CertificateDomainNames(cert *x509.Certificate) []string {
	names := []string{}
	names = append(names, cert.DNSNames...)
	names = append(names, cert.Subject.CommonName)
	return names
}
