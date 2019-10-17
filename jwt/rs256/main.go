package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/k0kubun/pp"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"io/ioutil"
	"time"
)

func readPrivateKey(path string) (*rsa.PrivateKey, error) {
	privateKeyData, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	privateKeyBlock, _ := pem.Decode(privateKeyData)
	if privateKeyBlock == nil {
		panic(errors.New("invalid private key data"))
	}
	if privateKeyBlock.Type != "RSA PRIVATE KEY" {
		panic(errors.New(fmt.Sprintf("invalid private key type : %s", privateKeyBlock.Type)))
	}

	return x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
}

func readCertificate(path string) (*x509.Certificate, error) {
	certificateData, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	certificateBlock, _ := pem.Decode(certificateData)
	if certificateBlock == nil {
		panic(errors.New("invalid private key data"))
	}
	return x509.ParseCertificate(certificateBlock.Bytes)
}

func main() {
	privateKey, err := readPrivateKey("./service-account-key.pem")
	if err != nil {
		panic(err)
	}

	privateJWK := &jose.JSONWebKey{
		Algorithm: string(jose.RS256),
		Key:       privateKey,
		Use:       "sig",
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{
			Key:       privateJWK,
			Algorithm: jose.RS256,
		},
		nil,
	)

	cl := jwt.Claims{
		Subject:   "subject",
		Issuer:    "issuer",
		NotBefore: jwt.NewNumericDate(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)),
		Audience:  jwt.Audience{"leela", "fry"},
		ID:        "ryo",
	}
	raw, err := jwt.Signed(signer).Claims(cl).CompactSerialize()
	if err != nil {
		panic(err)
	}

	fmt.Println(raw)

	tok, err := jwt.ParseSigned(raw)
	if err != nil {
		panic(err)
	}

	certificate, err := readCertificate("./service-account.pem")
	if err != nil {
		panic(err)
	}

	out := jwt.Claims{}
	err = tok.Claims(certificate.PublicKey, &out)
	if err != nil {
		panic(err)
	}
	pp.Print(out)

}
