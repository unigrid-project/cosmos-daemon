package types

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	jsonString string = "{" +
		"\"timeStamp\":\"2023-01-01T18:39:22.036Z\"," +
		"\"previousTimeStamp\":\"2021-01-01T18:38:17.183Z\"," +
		"\"data\":{" +
		"\"parameters\":{" +
		"\"vesting\":{\"denom\":\"UGD\",\"amount\":\"123\"}," +
		"\"genesisTransactions\":{\"rate\":\"333\",\"maxRate\":\"3333331\"}," +
		"\"consensusBlock\":{\"maxBytes\":\"2222\",\"maxGas\":\"22222222\"}" +
		"}}}"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
)

func serverSetup() func() {
	mux = http.NewServeMux()

	// priv, err := rsa.GenerateKey(rand.Reader, *rsaBits)
	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)

	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	out := &bytes.Buffer{}
	out2 := &bytes.Buffer{}
	pem.Encode(out, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certKey := out.Bytes()

	out.Reset()
	pem.Encode(out2, pemBlockForKey(priv))
	pubKey := out2.Bytes()

	cert, err := tls.X509KeyPair(certKey, pubKey)

	if err != nil {
		log.Panic("bad server certs: ", err)
	}
	certs := []tls.Certificate{cert}

	server = httptest.NewUnstartedServer(mux)
	server.TLS = &tls.Config{Certificates: certs}
	//server.URL = "http://127.0.0.1:52884"
	//server. = "0.0.0.0:52884"
	server.StartTLS()

	return func() {
		server.Close()
	}
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func TestSavingParamsFromHedgehog(t *testing.T) {
	teardown := serverSetup()
	defer teardown()

	mux.HandleFunc("/gridspork/cosmos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(jsonString))
	})

	cache := NewCache()

	require.Equal(t, len(cache.params), 0)

	cache.callHedgehog(server.URL + "/gridspork/cosmos")

	// data from jsonString variable
	mockResponse := make(map[string]UgdParam)
	mockResponse["parameters"] = UgdParam{
		Vesting{Denom: "UGD", Amount: "123"},
		ConsensusBlock{MaxBytes: "2222", MaxGas: "22222222"},
		GenesisTransactions{Rate: "333", MaxRate: "3333331"},
	}

	require.NotEmpty(t, len(cache.params))
	require.Equal(t, mockResponse, cache.params)
}
