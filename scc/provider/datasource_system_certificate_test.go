package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSystemCertificate(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_system_certficate")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceSystemCertificate("scc_sys_cert"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr("data.scc_system_certificate.scc_sys_cert", "subject_dn", regexp.MustCompile(`CN=.*?(,.*)?`)),
						resource.TestMatchResourceAttr("data.scc_system_certificate.scc_sys_cert", "issuer", regexp.MustCompile(`CN=.*?(,.*)?`)),
						resource.TestMatchResourceAttr("data.scc_system_certificate.scc_sys_cert", "valid_from", regexValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_system_certificate.scc_sys_cert", "valid_to", regexValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_system_certificate.scc_sys_cert", "serial_number", regexValidSerialNumber),
						resource.TestCheckResourceAttrSet("data.scc_system_certificate.scc_sys_cert", "certificate_pem"),
					),
				},
			},
		})

	})

}

func DataSourceSystemCertificate(datasourceName string) string {
	return fmt.Sprintf(`data "scc_system_certificate" "%s" {}`, datasourceName)
}
