package api

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestApiClient_BasicAuth(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "testuser" || password != "testpassword" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "success"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client, diags := createBasicAuthClient(server.URL)
	if diags.HasError() {
		t.Fatalf("failed to create basic auth client: %v", diags)
	}

	t.Run("GET /success", func(t *testing.T) {
		resp, diags := client.GetRequest("/success")
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", resp.StatusCode)
		}
	})
}

func TestRestApiClient_CertificateAuth(t *testing.T) {
	// Generate server cert
	serverCertPEM, serverKeyPEM, _, diags := generateSelfSignedCert()
	if diags.HasError() {
		t.Fatalf("server cert generation failed: %v", diags)
	}

	// Generate client cert
	clientCertPEM, clientKeyPEM, clientCert, diags := generateSelfSignedCert()
	if diags.HasError() {
		t.Fatalf("client cert generation failed: %v", diags)
	}

	// Create server that validates client cert
	clientCertPool := x509.NewCertPool()
	clientCertPool.AddCert(clientCert)

	serverTLSCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		t.Fatalf("invalid server cert/key: %v", err)
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/secure", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "secured"}`))
	})

	server := httptest.NewUnstartedServer(handler)
	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{serverTLSCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCertPool,
	}
	server.StartTLS()
	defer server.Close()

	client, diags := createCertAuthClient(server.URL, serverCertPEM, clientCertPEM, clientKeyPEM)
	if diags.HasError() {
		t.Fatalf("failed to create cert auth client: %v", diags)
	}

	t.Run("GET /secure", func(t *testing.T) {
		resp, diags := client.GetRequest("/secure")
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", resp.StatusCode)
		}
	})
}

func TestRestApiClient_BothAuthProvidedFails(t *testing.T) {
	baseURL, _ := url.Parse("https://localhost")
	certPEM, keyPEM, _, _ := generateSelfSignedCert()
	// Provided both basic authentication(username/password) and certificate based authentication to the function
	client, diags := NewRestApiClient(nil, baseURL, "user", "pass", certPEM, certPEM, keyPEM)

	assert.Nil(t, client, "expected client to be nil when both auth methods are provided")
	assert.True(t, diags.HasError(), "expected diagnostics to have error")

	require.Len(t, diags, 1)
	assert.Contains(t, diags[0].Summary(), "Authentication")
	assert.Contains(t, diags[0].Detail(), "Cannot use both certificate-based and basic authentication simultaneously. Please choose one method.")
}

func TestRestApiClient_NoAuthProvidedFails(t *testing.T) {
	baseURL, _ := url.Parse("https://localhost")
	// Provided neither basic authentication(username/password) nor certificate based authentication to the function
	client, diags := NewRestApiClient(nil, baseURL, "", "", nil, nil, nil)
	assert.Nil(t, client)
	assert.True(t, diags.HasError(), "expected error for no auth provided")

	require.Len(t, diags, 1)
	assert.Contains(t, diags[0].Summary(), "Authentication")
	assert.Contains(t, diags[0].Detail(), "Either certificate-based or basic authentication must be provided.")
}

func TestRestApiClient_InvalidClientCertFails(t *testing.T) {
	baseURL, _ := url.Parse("https://localhost")
	// Generate invalid client certificate and key and provided to the function
	invalidPEM := []byte("not a valid pem")
	client, diags := NewRestApiClient(nil, baseURL, "", "", nil, invalidPEM, invalidPEM)

	assert.Nil(t, client)
	assert.True(t, diags.HasError(), "expected error for invalid client cert")

	require.Len(t, diags, 1)
	assert.Contains(t, diags[0].Summary(), "Missing CA Certificate")
	assert.Contains(t, diags[0].Detail(), "When using certificate-based authentication, a CA certificate must be provided.")
}

func TestRestApiClient_InvalidCACertFails(t *testing.T) {
	baseURL, _ := url.Parse("https://localhost")
	// Generate valid client certificate and key
	certPEM, keyPEM, _, _ := generateSelfSignedCert()
	// Generate invalid CA Certificate
	invalidCA := []byte("not valid pem")
	client, diags := NewRestApiClient(nil, baseURL, "", "", invalidCA, certPEM, keyPEM)

	assert.Nil(t, client)
	assert.True(t, diags.HasError(), "expected CA cert parse error")

	require.Len(t, diags, 1)
	assert.Contains(t, diags[0].Summary(), "CA Certificate")
	assert.Contains(t, diags[0].Detail(), "The provided CA certificate is not valid PEM-encoded data.")
}

// generateSelfSignedCert generates a self-signed TLS certificate and its private key.
func generateSelfSignedCert() (certPEM, keyPEM []byte, cert *x509.Certificate, diags diag.Diagnostics) {
	// Generate a new RSA private key with 2048-bit length
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		diags.AddError("Private Key Generation Failed", fmt.Sprintf("failed to generate private key: %v", err))
		return nil, nil, nil, diags
	}

	// Create a certificate template with required fields
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,

		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:    []string{"localhost"},
	}

	// Create a self-signed certificate using the template and the generated private key.
	// The template is being used for parent issuer certificate and the certificate itself since it is a self-signed certfificate.
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		diags.AddError("Certificate Creation Failed", fmt.Sprintf("failed to create certificate: %v", err))
		return nil, nil, nil, diags
	}

	// Encode the certificate & private key to PEM format
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)})

	// Parse the DER-encoded certificate into an x509.Certificate object
	parsedCert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		diags.AddError("Certificate Parsing Failed", fmt.Sprintf("failed to parse certificate: %v", err))
		return nil, nil, nil, diags
	}

	// Return the PEM-encoded certificate, key, and parsed certificate
	return certPEM, keyPEM, parsedCert, diags
}

func TestValidateResponse_SuccessCodes(t *testing.T) {
	successStatuses := []int{http.StatusOK, http.StatusCreated, http.StatusNoContent}

	for _, code := range successStatuses {
		resp := &http.Response{
			StatusCode: code,
			Body:       io.NopCloser(bytes.NewBuffer(nil)),
		}

		diags := validateResponse(resp)
		if diags != nil {
			t.Errorf("expected no error for status %d, got %v", code, diags)
		}
	}
}

func TestValidateResponse_ErrorWithValidJSON(t *testing.T) {
	body := `{"type":"BadRequest","message":"Invalid input"}`
	req, _ := http.NewRequest(http.MethodPost, "http://example.com/api", nil)

	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Request:    req,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}

	diags := validateResponse(resp)
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics, got none")
	}

	got := diags[0].Summary() + ": " + diags[0].Detail()
	expected := "API Error: HTTP POST http://example.com/api failed with status 400: Invalid input"

	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestValidateResponse_ErrorWithInvalidJSON(t *testing.T) {
	body := `<<<garbage>>>`
	req, _ := http.NewRequest(http.MethodDelete, "http://example.com/delete", nil)

	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Request:    req,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}

	diags := validateResponse(resp)
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics, got none")
	}

	gotSummary := diags[0].Summary()
	expectedSummary := "API Error"
	if gotSummary != expectedSummary {
		t.Errorf("expected summary %q, got %q", expectedSummary, gotSummary)
	}

	if !strings.Contains(diags[0].Detail(), "Raw response: <<<garbage>>>") {
		t.Errorf("expected raw response detail, got %q", diags[0].Detail())
	}
}

func createBasicAuthClient(serverURL string) (*RestApiClient, diag.Diagnostics) {
	var diags diag.Diagnostics
	baseURL, err := url.Parse(serverURL)
	if err != nil {
		diags.AddError("Invalid URL", fmt.Sprintf("invalid server URL: %v", diags))
		return nil, diags
	}

	return NewRestApiClient(nil, baseURL, "testuser", "testpassword", nil, nil, nil)
}

func createCertAuthClient(serverURL string, serverCACert, clientCert, clientKey []byte) (*RestApiClient, diag.Diagnostics) {
	var diags diag.Diagnostics
	baseURL, err := url.Parse(serverURL)
	if err != nil {
		diags.AddError("Invalid URL", fmt.Sprintf("invalid server URL: %v", diags))
		return nil, diags
	}

	return NewRestApiClient(nil, baseURL, "", "", serverCACert, clientCert, clientKey)
}

func TestRestApiClient_DoRequest_BinaryResponse(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/cert", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pkix-cert")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte{0x01, 0x02, 0x03})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client, diags := createBasicAuthClient(server.URL)
	require.False(t, diags.HasError())

	resp, diags := client.DoRequest(http.MethodGet, "/cert", nil, "application/pkix-cert")

	require.False(t, diags.HasError())
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, body)
}

