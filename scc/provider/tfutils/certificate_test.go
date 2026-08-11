package tfutils_test

import (
	"context"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/resources"
	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTestClient(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := tfutils.NewTestClient(t, srv)

	require.NotNil(t, client)
	assert.Equal(t, srv.URL, client.BaseURL.String())
	assert.Equal(t, srv.Client(), client.Client)
}

func TestGenerateTestCert(t *testing.T) {
	t.Parallel()

	certPEM := tfutils.GenerateTestCert(t)

	require.NotEmpty(t, certPEM)
	block, rest := pem.Decode([]byte(certPEM))
	require.NotNil(t, block, "expected valid PEM block")
	assert.Equal(t, "CERTIFICATE", block.Type)
	assert.Empty(t, rest, "expected no trailing data after PEM block")
}

func TestGenerateValidDERCert(t *testing.T) {
	t.Parallel()

	der := tfutils.GenerateValidDERCert(t)

	require.NotEmpty(t, der)
}

func TestBuildSignedChainPlan_WithoutSAN(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := resources.NewSystemCertificateSignedChainResource()
	chain := tfutils.GenerateTestCert(t)

	plan := tfutils.BuildSignedChainPlan(ctx, r, chain, false)

	assert.NotNil(t, plan.Raw)
	assert.False(t, plan.Raw.IsNull())
}

func TestBuildSignedChainPlan_WithSAN(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := resources.NewUICertificateSignedChainResource()
	chain := tfutils.GenerateTestCert(t)

	plan := tfutils.BuildSignedChainPlan(ctx, r, chain, true)

	assert.NotNil(t, plan.Raw)
	assert.False(t, plan.Raw.IsNull())
}

func TestBuildSignedChainState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := resources.NewSystemCertificateSignedChainResource().(*resources.SystemCertificateSignedChainResource)
	chain := tfutils.GenerateTestCert(t)

	state := tfutils.BuildSignedChainState(ctx, r, chain)

	assert.NotNil(t, state.Raw)
	assert.False(t, state.Raw.IsNull())
}
