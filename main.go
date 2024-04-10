package main

// The code below has been derived from the following project:
//  https://github.com/containers/image/
// Specifically:
//  https://github.com/containers/image/blob/34a882db692a29586d1acd925adcf046d06175c4/signature/sigstore/generate.go
//  https://github.com/containers/image/blob/34a882db692a29586d1acd925adcf046d06175c4/signature/sigstore/copied.go

// Code from containers/image is licensed under Apache 2.0 from Apache contributors

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"syscall/js"

	"github.com/secure-systems-lab/go-securesystemslib/encrypted"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
)

const (
	// from sigstore/cosign/pkg/cosign.CosignPrivateKeyPemType.
	cosignPrivateKeyPemType = "ENCRYPTED COSIGN PRIVATE KEY"
	// from sigstore/cosign/pkg/cosign.SigstorePrivateKeyPemType.
	sigstorePrivateKeyPemType = "ENCRYPTED SIGSTORE PRIVATE KEY"
)

// simplified from sigstore/cosign/pkg/cosign.marshalKeyPair
// loadPrivateKey always requires a encryption, so this always requires a passphrase.
func marshalKeyPair(privateKey crypto.PrivateKey, publicKey crypto.PublicKey, password []byte) (_privateKey []byte, _publicKey []byte, err error) {
	x509Encoded, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("x509 encoding private key: %w", err)
	}

	encBytes, err := encrypted.Encrypt(x509Encoded, password)
	if err != nil {
		return nil, nil, err
	}

	// store in PEM format
	privBytes := pem.EncodeToMemory(&pem.Block{
		Bytes: encBytes,
		// Use the older “COSIGN” type name; as of 2023-03-30 cosign’s main branch generates “SIGSTORE” types,
		// but a version of cosign that can accept them has not yet been released.
		Type: sigstorePrivateKeyPemType,
	})

	// Now do the public key
	pubBytes, err := cryptoutils.MarshalPublicKeyToPEM(publicKey)
	if err != nil {
		return nil, nil, err
	}

	return privBytes, pubBytes, nil
}

// GenerateKeyPairResult is a struct to ensure the private and public parts can not be confused by the caller.
type GenerateKeyPairResult struct {
	PublicKey  []byte
	PrivateKey []byte
}

// GenerateKeyPair generates a public/private key pair usable for signing images using the sigstore format,
// and returns key representations suitable for storing in long-term files (with the private key encrypted using the provided passphrase).
// The specific key kind (e.g. algorithm, size), as well as the file format, are unspecified by this API,
// and can change with best practices over time.
func GenerateKeyPair(passphrase []byte) (*GenerateKeyPairResult, error) {
	// https://github.com/sigstore/cosign/blob/main/specs/SIGNATURE_SPEC.md#signature-schemes
	// only requires ECDSA-P256 to be supported, so that’s what we must use.
	rawKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		// Coverage: This can fail only if the randomness source fails
		return nil, err
	}
	private, public, err := marshalKeyPair(rawKey, rawKey.Public(), passphrase)
	if err != nil {
		return nil, err
	}
	return &GenerateKeyPairResult{
		PublicKey: public,
		PrivateKey: private,
	}, nil
}

// This is an original function not from containers/image
func main() {
 	keypair, err := GenerateKeyPair([]byte(""))
	 if err != nil {
		fmt.Errorf("%w", err)
	}
	// sets the generated keys as global variables in javascript
	js.Global().Set("cosignPublicKey", string(keypair.PublicKey))
	js.Global().Set("cosignPrivateKey", string(keypair.PrivateKey))
}