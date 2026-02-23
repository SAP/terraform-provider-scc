package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResourceSystemCertificateSelfSigned(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_system_certificate_self_signed")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSystemCertificateSelfSigned("scc_sys_cert_ss", 2048, "testCertificate"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_system_certificate_self_signed.scc_sys_cert_ss", "subject_dn.cn", "testCertificate"),
						resource.TestMatchResourceAttr("scc_system_certificate_self_signed.scc_sys_cert_ss", "issuer", regexp.MustCompile(`CN=.*?(,.*)?`)),
						resource.TestMatchResourceAttr("scc_system_certificate_self_signed.scc_sys_cert_ss", "valid_from", regexValidTimeStamp),
						resource.TestMatchResourceAttr("scc_system_certificate_self_signed.scc_sys_cert_ss", "valid_to", regexValidTimeStamp),
						resource.TestMatchResourceAttr("scc_system_certificate_self_signed.scc_sys_cert_ss", "serial_number", regexValidSerialNumber),
						resource.TestCheckResourceAttrSet("scc_system_certificate_self_signed.scc_sys_cert_ss", "certificate_pem"),
					),
				},
			},
		})
	})

	t.Run("error path - key size mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSystemCertificateSelfSignedWoKeySize("scc_sys_cert_ss", "testCertificate"),
					ExpectError: regexp.MustCompile(`The argument "key_size" is required, but no definition was found.`),
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
					Config:      ResourceSystemCertificateSelfSignedWoSubjectDN("scc_sys_cert_ss", 2048),
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
					Config:      ResourceSystemCertificateSelfSignedWoCommonName("scc_sys_cert_ss", 2048),
					ExpectError: regexp.MustCompile(`Inappropriate value for attribute "subject_dn": attribute "cn" is required.`),
				},
			},
		})
	})

}

func ResourceSystemCertificateSelfSigned(datasourceName string, keySize int64, commonName string) string {
	return fmt.Sprintf(`
	resource "scc_system_certificate_self_signed" "%s" {
  		key_size = %d
  		subject_dn = {
    		cn = "%s"
  		}
	}`, datasourceName, keySize, commonName)
}

func ResourceSystemCertificateSelfSignedWoKeySize(datasourceName string, commonName string) string {
	return fmt.Sprintf(`
	resource "scc_system_certificate_self_signed" "%s" {
  		subject_dn = {
    		cn = "%s"
  		}
	}`, datasourceName, commonName)
}

func ResourceSystemCertificateSelfSignedWoSubjectDN(datasourceName string, keySize int64) string {
	return fmt.Sprintf(`
	resource "scc_system_certificate_self_signed" "%s" {
  		key_size = %d
	}`, datasourceName, keySize)
}

func ResourceSystemCertificateSelfSignedWoCommonName(datasourceName string, keySize int64) string {
	return fmt.Sprintf(`
	resource "scc_system_certificate_self_signed" "%s" {
		key_size = %d
  		subject_dn = {}
	}`, datasourceName, keySize)
}
