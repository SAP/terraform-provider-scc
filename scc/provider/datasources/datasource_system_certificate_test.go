package datasources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSystemCertificate(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/datasource_system_certficate")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: tfutils.ProviderConfig(user) + DataSourceSystemCertificate("scc_sys_cert"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scc_system_certificate.scc_sys_cert", "subject_dn.cn"),
						resource.TestMatchResourceAttr("data.scc_system_certificate.scc_sys_cert", "issuer", regexp.MustCompile(`CN=.*?(,.*)?`)),
						resource.TestMatchResourceAttr("data.scc_system_certificate.scc_sys_cert", "valid_from", tfutils.RegexpValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_system_certificate.scc_sys_cert", "valid_to", tfutils.RegexpValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_system_certificate.scc_sys_cert", "serial_number", tfutils.RegexpValidSerialNumber),
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
