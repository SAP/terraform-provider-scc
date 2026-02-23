package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceSubaccount(t *testing.T) {
	subaccount := "b1799d1c-ce91-4cd4-8b6e-dc8f4eaf0ad9"
	regionHost := "cf.eu12.hana.ondemand.com"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount")
		if len(user.CloudUsername) == 0 || len(user.CloudPassword) == 0 {
			t.Fatalf("Missing TF_VAR_cloud_user or TF_VAR_cloud_password for recording test fixtures")
		}
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccount("scc_sa", regionHost, subaccount, user.CloudUsername, user.CloudPassword, "subaccount added via terraform tests"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "region_host", regionHost),
						resource.TestMatchResourceAttr("scc_subaccount.scc_sa", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "cloud_user", user.CloudUsername),
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "cloud_password", user.CloudPassword),
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "description", "subaccount added via terraform tests"),
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "location_id", ""),
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "auto_renew_before_days", "14"),

						resource.TestMatchResourceAttr("scc_subaccount.scc_sa", "tunnel.connected_since", regexValidTimeStamp),
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "tunnel.connections", "0"),
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "tunnel.state", "Connected"),
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "tunnel.user", user.CloudUsername),

						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "tunnel.application_connections.#", "0"),
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "tunnel.service_channels.#", "0"),

						resource.TestMatchResourceAttr("scc_subaccount.scc_sa", "tunnel.subaccount_certificate.issuer", regexp.MustCompile(`CN=.*?,OU=.*?,O=.*?,L=.*?,C=.*?`)),
						resource.TestMatchResourceAttr("scc_subaccount.scc_sa", "tunnel.subaccount_certificate.valid_to", regexValidTimeStamp),
						resource.TestMatchResourceAttr("scc_subaccount.scc_sa", "tunnel.subaccount_certificate.valid_from", regexValidTimeStamp),
						resource.TestMatchResourceAttr("scc_subaccount.scc_sa", "tunnel.subaccount_certificate.serial_number", regexValidSerialNumber),
						resource.TestMatchResourceAttr("scc_subaccount.scc_sa", "tunnel.subaccount_certificate.subject_dn", regexp.MustCompile(`CN=.*?,L=.*?,OU=.*?,OU=.*?,O=.*?,C=.*?`)),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subaccount.scc_sa",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					ResourceName:                         "scc_subaccount.scc_sa",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateIdFunc:                    getImportStateForSubaccount("scc_subaccount.scc_sa"),
					ImportStateVerifyIdentifierAttribute: "subaccount",
					ImportStateVerifyIgnore: []string{
						"cloud_user",
						"cloud_password",
						"connected",
						"auto_renew_before_days",
					},
				},
				{
					ResourceName:  "scc_subaccount.scc_sa",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.comb1799d1c-ce91-4cd4-8b6e-dc8f4eaf0ad9", // malformed ID
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*subaccount.*Got:`),
				},
				{
					ResourceName:  "scc_subaccount.scc_sa",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.com,b1799d1c-ce91-4cd4-8b6e-dc8f4eaf0ad9,extra",
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*subaccount.*Got:`),
				},
			},
		})

	})

	t.Run("update path - update description and display name", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_update")
		if len(user.CloudUsername) == 0 || len(user.CloudPassword) == 0 {
			t.Fatalf("Missing TF_VAR_cloud_user or TF_VAR_cloud_password for recording test fixtures")
		}
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccountUpdateWithDisplayName("scc_sa", regionHost, subaccount, user.CloudUsername, user.CloudPassword, "Initial description", "Initial Display Name"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "description", "Initial description"),
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "display_name", "Initial Display Name"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subaccount.scc_sa",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact(regionHost),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				// Update with mismatched configuration should throw error
				{
					Config:      providerConfig(user) + ResourceSubaccountUpdateWithDisplayName("scc_sa", "cf.us10.hana.ondemand.com", subaccount, user.CloudUsername, user.CloudPassword, "Initial description", "Initial Display Name"),
					ExpectError: regexp.MustCompile(`(?is)failed to update the cloud connector subaccount due to mismatched\s+configuration values`),
				},
				{
					Config: providerConfig(user) + ResourceSubaccountUpdateWithDisplayName("scc_sa", regionHost, subaccount, user.CloudUsername, user.CloudPassword, "Updated description", "Updated Display Name"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "description", "Updated description"),
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "display_name", "Updated Display Name"),
					),
				},
			},
		})
	})

	t.Run("update path - tunnel state change", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_update_tunnel")
		if user.CloudUsername == "" || user.CloudPassword == "" {
			t.Fatalf("Missing TF_VAR_cloud_user or TF_VAR_cloud_password for recording test fixtures")
		}
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccountWithTunnelState("scc_sa", regionHost, subaccount, user.CloudUsername, user.CloudPassword, "Testing tunnel connected", true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "tunnel.state", "Connected"),
					),
				},
				// Update with mismatched configuration should throw error
				{
					Config:      providerConfig(user) + ResourceSubaccountWithTunnelState("scc_sa", "cf.us10.hana.ondemand.com", subaccount, user.CloudUsername, user.CloudPassword, "Testing tunnel disconnected", false),
					ExpectError: regexp.MustCompile(`(?is)failed to update the cloud connector subaccount due to mismatched\s+configuration values`),
				},
				{
					Config: providerConfig(user) + ResourceSubaccountWithTunnelState("scc_sa", regionHost, subaccount, user.CloudUsername, user.CloudPassword, "Testing tunnel disconnected", false),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "tunnel.state", "Disconnected"),
					),
				},
				{
					Config: providerConfig(user) + ResourceSubaccountWithTunnelState("scc_sa", regionHost, subaccount, user.CloudUsername, user.CloudPassword, "Testing tunnel reconnected", true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount.scc_sa", "tunnel.state", "Connected"),
					),
				},
			},
		})
	})

	t.Run("error path - region host mandatory", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_err_wo_region_host")

		if len(user.CloudUsername) == 0 || len(user.CloudPassword) == 0 {
			t.Fatalf("Missing TF_VAR_cloud_user or TF_VAR_cloud_password for recording test fixtures")
		}
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config:      providerConfig(user) + ResourceSubaccountWoRegionHost("scc_sa", subaccount, user.CloudUsername, user.CloudPassword, "subaccount added via terraform tests"),
					ExpectError: regexp.MustCompile(`The argument "region_host" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - subaccount id mandatory", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_err_wo_subaccount_id")
		if len(user.CloudUsername) == 0 || len(user.CloudPassword) == 0 {
			t.Fatalf("Missing TF_VAR_cloud_user or TF_VAR_cloud_password for recording test fixtures")
		}
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config:      providerConfig(user) + ResourceSubaccountWoID("scc_sa", regionHost, user.CloudUsername, user.CloudPassword, "subaccount added via terraform tests"),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - cloud_user mandatory", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_err_wo_cloud_user")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccountWoCloudUser("scc_sa", regionHost, subaccount, user.CloudPassword, "missing cloud user"),
					ExpectError: regexp.MustCompile(
						`Missing required credentials`,
					),
				},
			},
		})
	})

	t.Run("error path - cloud_password mandatory", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_err_wo_cloud_password")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccountWoCloudPassword("scc_sa", regionHost, subaccount, user.CloudPassword, "missing cloud password"),
					ExpectError: regexp.MustCompile(
						`Missing required credentials`,
					),
				},
			},
		})
	})
}

func ResourceSubaccount(datasourceName string, regionHost string, subaccount string, cloudUser string, cloudPassword string, description string) string {
	return fmt.Sprintf(`
	resource "scc_subaccount" "%s" {
    region_host= "%s"
    subaccount= "%s"
    cloud_user= "%s"
    cloud_password= "%s" 
    description= "%s"
	}
	`, datasourceName, regionHost, subaccount, cloudUser, cloudPassword, description)
}

func ResourceSubaccountWoRegionHost(datasourceName string, subaccount string, cloudUser string, cloudPassword string, description string) string {
	return fmt.Sprintf(`
	resource "scc_subaccount" "%s" {
    subaccount= "%s"
    cloud_user= "%s"
    cloud_password= "%s" 
    description= "%s"
	}
	`, datasourceName, subaccount, cloudUser, cloudPassword, description)
}

func ResourceSubaccountWoID(datasourceName string, regionHost string, cloudUser string, cloudPassword string, description string) string {
	return fmt.Sprintf(`
	resource "scc_subaccount" "%s" {
    region_host= "%s"
    cloud_user= "%s"
    cloud_password= "%s" 
    description= "%s"
	}
	`, datasourceName, regionHost, cloudUser, cloudPassword, description)
}

func ResourceSubaccountWoCloudUser(datasourceName, regionHost, subaccount, cloudPassword, description string) string {
	return fmt.Sprintf(`
	resource "scc_subaccount" "%s" {
		region_host    = "%s"
		subaccount     = "%s"
		cloud_password = "%s"
		description    = "%s"
	}
	`, datasourceName, regionHost, subaccount, cloudPassword, description)
}

func ResourceSubaccountWoCloudPassword(datasourceName, regionHost, subaccount, cloudUser, description string) string {
	return fmt.Sprintf(`
	resource "scc_subaccount" "%s" {
		region_host = "%s"
		subaccount  = "%s"
		cloud_user  = "%s"
		description = "%s"
	}
	`, datasourceName, regionHost, subaccount, cloudUser, description)
}

func ResourceSubaccountUpdateWithDisplayName(datasourceName, regionHost, subaccount, cloudUser, cloudPassword, description, displayName string) string {
	return fmt.Sprintf(`
resource "scc_subaccount" "%s" {
  region_host   = "%s"
  subaccount    = "%s"
  cloud_user    = "%s"
  cloud_password = "%s"
  description   = "%s"
  display_name  = "%s"
}
`, datasourceName, regionHost, subaccount, cloudUser, cloudPassword, description, displayName)
}

func ResourceSubaccountWithTunnelState(datasourceName, regionHost, subaccount, cloudUser, cloudPassword, description string, connected bool) string {
	return fmt.Sprintf(`
resource "scc_subaccount" "%s" {
  region_host    = "%s"
  subaccount     = "%s"
  cloud_user     = "%s"
  cloud_password = "%s"
  description    = "%s"
  connected = %t
}
`, datasourceName, regionHost, subaccount, cloudUser, cloudPassword, description, connected)
}

func getImportStateForSubaccount(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s",
			rs.Primary.Attributes["region_host"],
			rs.Primary.Attributes["subaccount"],
		), nil
	}
}
