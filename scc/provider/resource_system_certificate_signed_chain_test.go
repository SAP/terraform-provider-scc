package provider

import (
	"context"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

func TestSystemCertificateSignedChain_Metadata(t *testing.T) {
	r := NewSystemCertificateSignedChainResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "scc",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "scc_system_certificate_signed_chain", resp.TypeName)
}

func TestSystemCertificateSignedChain_Schema(t *testing.T) {
	r := NewSystemCertificateSignedChainResource()

	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema.Attributes["signed_chain"])
	assert.True(t, resp.Schema.Attributes["signed_chain"].IsRequired())
}

func TestSystemCertificateSignedChain_Configure_Success(t *testing.T) {
	r := NewSystemCertificateSignedChainResource().(*SystemCertificateSignedChainResource)

	client := &api.RestApiClient{}

	req := resource.ConfigureRequest{
		ProviderData: client,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Configure_WrongType(t *testing.T) {
	r := NewSystemCertificateSignedChainResource().(*SystemCertificateSignedChainResource)

	req := resource.ConfigureRequest{
		ProviderData: "wrong",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Create_InvalidPEM(t *testing.T) {
	// Invalid PEM should fail validation
	diags := validatePEMChain("not a cert")
	if diags == nil || !diags.HasError() {
		t.Fatalf("expected validation error for invalid PEM")
	}
}

func TestSystemCertificateSignedChain_Update(t *testing.T) {
	r := NewSystemCertificateSignedChainResource()

	resp := &resource.UpdateResponse{}
	r.Update(context.Background(), resource.UpdateRequest{}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Delete(t *testing.T) {
	r := &SystemCertificateSignedChainResource{
		client: &api.RestApiClient{},
	}

	// Mock API call
	old := requestAndUnmarshalFunc
	requestAndUnmarshalFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.SystemCertificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {
		return nil
	}
	defer func() { requestAndUnmarshalFunc = old }()

	resp := &resource.DeleteResponse{}
	req := resource.DeleteRequest{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestSystemCertificateSignedChain_Read(t *testing.T) {
	r := &SystemCertificateSignedChainResource{
		client: &api.RestApiClient{},
	}

	oldReq := requestAndUnmarshalFunc
	oldBin := getCertificateBinaryFunc

	requestAndUnmarshalFunc = func(
		client *api.RestApiClient,
		respObj *apiobjects.SystemCertificate,
		method string,
		endpoint string,
		body map[string]any,
		expectJSON bool,
	) diag.Diagnostics {
		return nil
	}

	getCertificateBinaryFunc = func(
		client *api.RestApiClient,
		endpoint string,
	) ([]byte, diag.Diagnostics) {
		return []byte("fakecert"), nil
	}

	defer func() {
		requestAndUnmarshalFunc = oldReq
		getCertificateBinaryFunc = oldBin
	}()

	resp := &resource.ReadResponse{}
	req := resource.ReadRequest{}

	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}
