package provider

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSubjectDN_AllFields(t *testing.T) {
	dn := "CN=testCert,EMAIL=test@example.com,L=Bangalore,OU=Engineering,O=SAP,ST=KA,C=IN"

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("test@example.com"), result.Email)
	assert.Equal(t, types.StringValue("Bangalore"), result.Locality)
	assert.Equal(t, types.StringValue("Engineering"), result.OrganizationalUnit)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
	assert.Equal(t, types.StringValue("KA"), result.State)
	assert.Equal(t, types.StringValue("IN"), result.Country)
}

func TestParseSubjectDN_MissingOptionalFields(t *testing.T) {
	dn := "CN=testCert,O=SAP"

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)

	assert.True(t, result.Email.IsNull())
	assert.True(t, result.Locality.IsNull())
	assert.True(t, result.OrganizationalUnit.IsNull())
	assert.True(t, result.State.IsNull())
	assert.True(t, result.Country.IsNull())
}

func TestParseSubjectDN_EmptyDN(t *testing.T) {
	result := parseSubjectDN("")

	assert.True(t, result.CommonName.IsNull())
	assert.True(t, result.Email.IsNull())
	assert.True(t, result.Locality.IsNull())
	assert.True(t, result.OrganizationalUnit.IsNull())
	assert.True(t, result.Organization.IsNull())
	assert.True(t, result.State.IsNull())
	assert.True(t, result.Country.IsNull())
}

func TestParseSubjectDN_UnknownFieldsIgnored(t *testing.T) {
	dn := "CN=testCert,XYZ=value,O=SAP"

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
}

func TestParseSubjectDN_CaseInsensitiveKeys(t *testing.T) {
	dn := "cn=testCert,o=SAP,c=in"

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
	assert.Equal(t, types.StringValue("in"), result.Country)
}

func TestParseSubjectDN_SpacesTrimmed(t *testing.T) {
	dn := " CN = testCert , O = SAP , C = IN "

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("testCert"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
	assert.Equal(t, types.StringValue("IN"), result.Country)
}

func TestParseSubjectDN_MultipleOU_LastWins(t *testing.T) {
	dn := "CN=testCert,OU=Team1,OU=Team2"

	result := parseSubjectDN(dn)

	// Current implementation keeps last OU
	assert.Equal(t, types.StringValue("Team2"), result.OrganizationalUnit)
}

func TestBuildSubjectDN_AllFields(t *testing.T) {
	subject := &CertificateSubjectDNConfig{
		CommonName:         types.StringValue("testCert"),
		Email:              types.StringValue("test@sap.com"),
		Locality:           types.StringValue("Bangalore"),
		OrganizationalUnit: types.StringValue("BTP"),
		Organization:       types.StringValue("SAP"),
		State:              types.StringValue("KA"),
		Country:            types.StringValue("IN"),
	}

	result := BuildSubjectDN(subject)

	expected := "CN=testCert,EMAIL=test@sap.com,L=Bangalore,OU=BTP,O=SAP,ST=KA,C=IN"
	assert.Equal(t, expected, result)
}

func TestBuildSubjectDN_OnlyCN(t *testing.T) {
	subject := &CertificateSubjectDNConfig{
		CommonName: types.StringValue("testCert"),
	}

	result := BuildSubjectDN(subject)

	assert.Equal(t, "CN=testCert", result)
}

func TestBuildSubjectDN_NilSubject(t *testing.T) {
	result := BuildSubjectDN(nil)
	assert.Equal(t, "", result)
}

func TestBuildSubjectDN_CNNull(t *testing.T) {
	subject := &CertificateSubjectDNConfig{
		CommonName: types.StringNull(),
	}

	result := BuildSubjectDN(subject)
	assert.Equal(t, "", result)
}

func TestBuildSubjectDN_EmptyOptionalFieldsIgnored(t *testing.T) {
	subject := &CertificateSubjectDNConfig{
		CommonName:         types.StringValue("testCert"),
		Email:              types.StringValue(""),
		Locality:           types.StringNull(),
		OrganizationalUnit: types.StringValue(" "),
		Organization:       types.StringValue("SAP"),
	}

	result := BuildSubjectDN(subject)

	assert.Equal(t, "CN=testCert,O=SAP", result)
}

func TestBuildSubjectDN_SpacesTrimmed(t *testing.T) {
	subject := &CertificateSubjectDNConfig{
		CommonName:   types.StringValue(" testCert "),
		Organization: types.StringValue(" SAP "),
		Country:      types.StringValue(" IN "),
	}

	result := BuildSubjectDN(subject)

	assert.Equal(t, "CN=testCert,O=SAP,C=IN", result)
}

func TestBuildSubjectDN_FieldOrder(t *testing.T) {
	subject := &CertificateSubjectDNConfig{
		CommonName:   types.StringValue("cert"),
		Country:      types.StringValue("IN"),
		Organization: types.StringValue("SAP"),
		Email:        types.StringValue("a@sap.com"),
		Locality:     types.StringValue("BLR"),
	}

	result := BuildSubjectDN(subject)

	expected := "CN=cert,EMAIL=a@sap.com,L=BLR,O=SAP,C=IN"
	assert.Equal(t, expected, result)
}

func TestBuildSubjectDN_OptionalFieldsNull(t *testing.T) {
	subject := &CertificateSubjectDNConfig{
		CommonName:   types.StringValue("cert"),
		Email:        types.StringNull(),
		Organization: types.StringNull(),
	}

	result := BuildSubjectDN(subject)

	assert.Equal(t, "CN=cert", result)
}

func TestBuildSubjectDN_SpecialCharacters(t *testing.T) {
	subject := &CertificateSubjectDNConfig{
		CommonName:   types.StringValue("test-cert_123"),
		Organization: types.StringValue("SAP-SE"),
	}

	result := BuildSubjectDN(subject)

	assert.Equal(t, "CN=test-cert_123,O=SAP-SE", result)
}

func TestBuildSubjectDN_RoundTrip(t *testing.T) {
	input := &CertificateSubjectDNConfig{
		CommonName:   types.StringValue("testCert"),
		Organization: types.StringValue("SAP"),
		Country:      types.StringValue("IN"),
	}

	dn := BuildSubjectDN(input)
	parsed := parseSubjectDN(dn)

	assert.Equal(t, input.CommonName, parsed.CommonName)
	assert.Equal(t, input.Organization, parsed.Organization)
	assert.Equal(t, input.Country, parsed.Country)
}

func TestParseSubjectDN_EmptyPartsIgnored(t *testing.T) {
	dn := "CN=test,,O=SAP, ,C=IN"

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("test"), result.CommonName)
	assert.Equal(t, types.StringValue("SAP"), result.Organization)
	assert.Equal(t, types.StringValue("IN"), result.Country)
}

func TestParseSubjectDN_DuplicateCN_LastWins(t *testing.T) {
	dn := "CN=one,CN=two"

	result := parseSubjectDN(dn)

	assert.Equal(t, types.StringValue("two"), result.CommonName)
}

func TestBuildSubjectDN_UnknownValuesIgnored(t *testing.T) {
	subject := &CertificateSubjectDNConfig{
		CommonName: types.StringValue("cert"),
		Email:      types.StringUnknown(),
	}

	result := BuildSubjectDN(subject)

	assert.Equal(t, "CN=cert", result)
}

func TestExpandSubjectDN_Null(t *testing.T) {
	ctx := context.Background()
	obj := types.ObjectNull(subjectDNAttrTypes.AttrTypes)

	res, diags := ExpandSubjectDN(ctx, obj)

	assert.Nil(t, res)
	assert.False(t, diags.HasError())
}
func TestExpandSubjectDN_Valid(t *testing.T) {
	ctx := context.Background()

	obj, diags := types.ObjectValue(
		subjectDNAttrTypes.AttrTypes,
		map[string]attr.Value{
			"cn":    types.StringValue("cert"),
			"email": types.StringNull(),
			"l":     types.StringNull(),
			"ou":    types.StringNull(),
			"o":     types.StringValue("SAP"),
			"st":    types.StringNull(),
			"c":     types.StringNull(),
		},
	)

	assert.False(t, diags.HasError())

	res, diags := ExpandSubjectDN(ctx, obj)

	assert.False(t, diags.HasError())
	assert.Equal(t, "cert", res.CommonName.ValueString())
	assert.Equal(t, "SAP", res.Organization.ValueString())
}

func TestValidatePEMData_Empty(t *testing.T) {
	diags := validatePEMData("")
	assert.True(t, diags.HasError())
}

func TestValidatePEMData_InvalidPEM(t *testing.T) {
	diags := validatePEMData("not a pem")
	assert.True(t, diags.HasError())
}

func TestValidatePEMData_UnsupportedType(t *testing.T) {
	pem := `-----BEGIN FOO-----
abcd
-----END FOO-----`

	diags := validatePEMData(pem)
	assert.True(t, diags.HasError())
}

func TestValidatePEMData_ValidCert(t *testing.T) {
	var validCert = generateTestCert(t)
	diags := validatePEMData(validCert)
	assert.False(t, diags.HasError())
}

func TestValidatePEMChain_Empty(t *testing.T) {
	diags := validatePEMChain("")
	assert.True(t, diags.HasError())
}

func TestValidatePEMChain_InvalidBlock(t *testing.T) {
	data := `-----BEGIN PRIVATE KEY-----
abcd
-----END PRIVATE KEY-----`

	diags := validatePEMChain(data)
	assert.True(t, diags.HasError())
}

func TestValidatePEMChain_MultipleCerts(t *testing.T) {
	var validCert = generateTestCert(t)
	data := validCert + "\n" + validCert
	diags := validatePEMChain(data)
	assert.False(t, diags.HasError())
}

func generateTestCert(t *testing.T) string {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		Subject: pkix.Name{
			CommonName: "test-cert",
		},
	}

	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&priv.PublicKey,
		priv,
	)
	require.NoError(t, err)

	var pemBuf bytes.Buffer
	err = pem.Encode(&pemBuf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})
	require.NoError(t, err)

	return pemBuf.String()
}