package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceCACertificate(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_ca_certficate")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceCACertificate("scc_ca_cert"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr("data.scc_ca_certificate.scc_ca_cert", "subject_dn", regexp.MustCompile(`CN=.*?(,.*)?`)),
						resource.TestMatchResourceAttr("data.scc_ca_certificate.scc_ca_cert", "issuer", regexp.MustCompile(`CN=.*?(,.*)?`)),
						resource.TestMatchResourceAttr("data.scc_ca_certificate.scc_ca_cert", "valid_from", regexValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_ca_certificate.scc_ca_cert", "valid_to", regexValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_ca_certificate.scc_ca_cert", "serial_number", regexValidSerialNumber),
						resource.TestCheckResourceAttrSet("data.scc_ca_certificate.scc_ca_cert", "certificate_pem"),
					),
				},
			},
		})

	})

}

func DataSourceCACertificate(datasourceName string) string {
	return fmt.Sprintf(`data "scc_ca_certificate" "%s" {}`, datasourceName)
}
