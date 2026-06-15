package resources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestResourceProxySettings(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/resource_proxy_settings")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: tfutils.ProviderConfig(user) + ResourceProxySettings("scc_ps", "testHost", 123, "testUser", "testPassword"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_proxy_settings.scc_ps", "host", "testHost"),
						resource.TestCheckResourceAttr("scc_proxy_settings.scc_ps", "port", "123"),
						resource.TestCheckResourceAttr("scc_proxy_settings.scc_ps", "user", "testUser"),
						resource.TestCheckResourceAttrSet("scc_proxy_settings.scc_ps", "password"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_proxy_settings.scc_ps",
							map[string]knownvalue.Check{
								"id": knownvalue.StringExact("proxy-settings"),
							},
						),
					},
				},
				{
					ResourceName:                         "scc_proxy_settings.scc_ps",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateId:                        "proxy-settings",
					ImportStateVerifyIdentifierAttribute: "id",
					ImportStateVerifyIgnore: []string{
						"password",
					},
				},
				{
					Config: tfutils.ProviderConfig(user) + ResourceProxySettings("scc_ps", "updatedTestHost", 123, "testUser", "testPassword"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_proxy_settings.scc_ps", "host", "updatedTestHost"),
						resource.TestCheckResourceAttr("scc_proxy_settings.scc_ps", "port", "123"),
						resource.TestCheckResourceAttr("scc_proxy_settings.scc_ps", "user", "testUser"),
						resource.TestCheckResourceAttrSet("scc_proxy_settings.scc_ps", "password"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_proxy_settings.scc_ps",
							map[string]knownvalue.Check{
								"id": knownvalue.StringExact("proxy-settings"),
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - host mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceProxySettingsWoHost("scc_ps", 123, "testUser", "testPassword"),
					ExpectError: regexp.MustCompile(`The argument "host" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - port mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceProxySettingsWoPort("scc_ps", "testHost", "testUser", "testPassword"),
					ExpectError: regexp.MustCompile(`The argument "port" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - user mandatory if password is set", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceProxySettingsWithOnlyPassword("scc_ps", "testHost", 123, "testPassword"),
					ExpectError: regexp.MustCompile(`Attribute "user" must be specified when "password" is specified`),
				},
			},
		})
	})

	t.Run("error path - invalid import id", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config: ResourceProxySettings(
						"scc_ps",
						"testHost",
						123,
						"testUser",
						"testPassword",
					),
				},
				{
					ResourceName:  "scc_proxy_settings.scc_ps",
					ImportState:   true,
					ImportStateId: "invalid",
					ExpectError:   regexp.MustCompile("Expected import identifier"),
					ImportStateVerifyIgnore: []string{
						"password",
					},
				},
			},
		})
	})
}

func ResourceProxySettings(resourceName string, host string, port int, user string, password string) string {
	return fmt.Sprintf(`
	resource "scc_proxy_settings" "%s" {
    host = "%s"
    user = "%s"
    password = "%s"
    port = %d
	}
	`, resourceName, host, user, password, port)
}

func ResourceProxySettingsWoHost(resourceName string, port int, user string, password string) string {
	return fmt.Sprintf(`
	resource "scc_proxy_settings" "%s" {
    port = %d
    user = "%s"
    password = "%s"
	}
	`, resourceName, port, user, password)
}

func ResourceProxySettingsWoPort(resourceName string, host, user, password string) string {
	return fmt.Sprintf(`
	resource "scc_proxy_settings" "%s" {
    host = "%s"
    user = "%s"
    password = "%s"
	}
	`, resourceName, host, user, password)
}

func ResourceProxySettingsWithOnlyPassword(resourceName string, host string, port int, password string) string {
	return fmt.Sprintf(`
	resource "scc_proxy_settings" "%s" {
    host = "%s"
    port = %d
    password = "%s"
	}
	`, resourceName, host, port, password)
}
