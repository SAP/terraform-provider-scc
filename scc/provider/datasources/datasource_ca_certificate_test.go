package datasources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceCACertificate(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/datasource_ca_certficate")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: tfutils.ProviderConfig(user) + DataSourceCACertificate("scc_ca_cert"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scc_ca_certificate.scc_ca_cert", "subject_dn.cn"),
						resource.TestMatchResourceAttr("data.scc_ca_certificate.scc_ca_cert", "issuer", regexp.MustCompile(`CN=.*?(,.*)?`)),
						resource.TestMatchResourceAttr("data.scc_ca_certificate.scc_ca_cert", "valid_from", tfutils.RegexpValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_ca_certificate.scc_ca_cert", "valid_to", tfutils.RegexpValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_ca_certificate.scc_ca_cert", "serial_number", tfutils.RegexpValidSerialNumber),
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
