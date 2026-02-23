package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
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
		CommonName: types.StringValue(" testCert "),
		Organization: types.StringValue(" SAP "),
		Country: types.StringValue(" IN "),
	}

	result := BuildSubjectDN(subject)

	assert.Equal(t, "CN=testCert,O=SAP,C=IN", result)
}

func TestBuildSubjectDN_FieldOrder(t *testing.T) {
	subject := &CertificateSubjectDNConfig{
		CommonName:         types.StringValue("cert"),
		Country:            types.StringValue("IN"),
		Organization:       types.StringValue("SAP"),
		Email:              types.StringValue("a@sap.com"),
		Locality:           types.StringValue("BLR"),
	}

	result := BuildSubjectDN(subject)

	expected := "CN=cert,EMAIL=a@sap.com,L=BLR,O=SAP,C=IN"
	assert.Equal(t, expected, result)
}

func TestBuildSubjectDN_OptionalFieldsNull(t *testing.T) {
	subject := &CertificateSubjectDNConfig{
		CommonName: types.StringValue("cert"),
		Email:      types.StringNull(),
		Organization: types.StringNull(),
	}

	result := BuildSubjectDN(subject)

	assert.Equal(t, "CN=cert", result)
}

func TestBuildSubjectDN_SpecialCharacters(t *testing.T) {
	subject := &CertificateSubjectDNConfig{
		CommonName: types.StringValue("test-cert_123"),
		Organization: types.StringValue("SAP-SE"),
	}

	result := BuildSubjectDN(subject)

	assert.Equal(t, "CN=test-cert_123,O=SAP-SE", result)
}

func TestBuildSubjectDN_RoundTrip(t *testing.T) {
	input := &CertificateSubjectDNConfig{
		CommonName:         types.StringValue("testCert"),
		Organization:       types.StringValue("SAP"),
		Country:            types.StringValue("IN"),
	}

	dn := BuildSubjectDN(input)
	parsed := parseSubjectDN(dn)

	assert.Equal(t, input.CommonName, parsed.CommonName)
	assert.Equal(t, input.Organization, parsed.Organization)
	assert.Equal(t, input.Country, parsed.Country)
}
