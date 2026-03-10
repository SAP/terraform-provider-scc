package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResourceCACertificateSelfSigned(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_ca_certificate_self_signed")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceCACertificateSelfSigned("scc_ca_cert_ss", 2048, "testCertificate"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_ca_certificate_self_signed.scc_ca_cert_ss", "subject_dn.cn", "testCertificate"),
						resource.TestMatchResourceAttr("scc_ca_certificate_self_signed.scc_ca_cert_ss", "issuer", regexp.MustCompile(`CN=.*?(,.*)?`)),
						resource.TestMatchResourceAttr("scc_ca_certificate_self_signed.scc_ca_cert_ss", "valid_from", regexValidTimeStamp),
						resource.TestMatchResourceAttr("scc_ca_certificate_self_signed.scc_ca_cert_ss", "valid_to", regexValidTimeStamp),
						resource.TestMatchResourceAttr("scc_ca_certificate_self_signed.scc_ca_cert_ss", "serial_number", regexValidSerialNumber),
						resource.TestCheckResourceAttrSet("scc_ca_certificate_self_signed.scc_ca_cert_ss", "certificate_pem"),
					),
				},
			},
		})
	})

	t.Run("error path - subject dn mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceCACertificateSelfSignedWoSubjectDN("scc_ca_cert_ss", 2048),
					ExpectError: regexp.MustCompile(`The argument "subject_dn" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - common name mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceCACertificateSelfSignedWoCommonName("scc_ca_cert_ss", 2048),
					ExpectError: regexp.MustCompile(`Inappropriate value for attribute "subject_dn": attribute "cn" is required.`),
				},
			},
		})
	})

}

func ResourceCACertificateSelfSigned(datasourceName string, keySize int64, commonName string) string {
	return fmt.Sprintf(`
	resource "scc_ca_certificate_self_signed" "%s" {
  		key_size = %d
  		subject_dn = {
    		cn = "%s"
  		}
	}`, datasourceName, keySize, commonName)
}

func ResourceCACertificateSelfSignedWoSubjectDN(datasourceName string, keySize int64) string {
	return fmt.Sprintf(`
	resource "scc_ca_certificate_self_signed" "%s" {
  		key_size = %d
	}`, datasourceName, keySize)
}

func ResourceCACertificateSelfSignedWoCommonName(datasourceName string, keySize int64) string {
	return fmt.Sprintf(`
	resource "scc_ca_certificate_self_signed" "%s" {
		key_size = %d
  		subject_dn = {}
	}`, datasourceName, keySize)
}
