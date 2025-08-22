package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceSubaccount(t *testing.T) {
	subaccountId := "4916a705-273c-45a6-a2f0-08c234c7a23d"
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
					Config: providerConfig(user) + ResourceSubaccount("test", regionHost, subaccountId, user.CloudUsername, user.CloudPassword, "subaccount added via terraform tests"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("scc_subaccount.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("scc_subaccount.test", "cloud_user", user.CloudUsername),
						resource.TestCheckResourceAttr("scc_subaccount.test", "cloud_password", user.CloudPassword),
						resource.TestCheckResourceAttr("scc_subaccount.test", "description", "subaccount added via terraform tests"),
						resource.TestCheckResourceAttr("scc_subaccount.test", "location_id", ""),

						resource.TestMatchResourceAttr("scc_subaccount.test", "tunnel.connected_since_time_stamp", regexValidTimeStamp),
						resource.TestCheckResourceAttr("scc_subaccount.test", "tunnel.connections", "0"),
						resource.TestCheckResourceAttr("scc_subaccount.test", "tunnel.state", "Connected"),
						resource.TestCheckResourceAttr("scc_subaccount.test", "tunnel.user", user.CloudUsername),

						resource.TestCheckResourceAttr("scc_subaccount.test", "tunnel.application_connections.#", "0"),
						resource.TestCheckResourceAttr("scc_subaccount.test", "tunnel.service_channels.#", "0"),

						resource.TestMatchResourceAttr("scc_subaccount.test", "tunnel.subaccount_certificate.issuer", regexp.MustCompile(`CN=.*?,OU=S.*?,O=.*?,L=.*?,C=.*?`)),
						resource.TestMatchResourceAttr("scc_subaccount.test", "tunnel.subaccount_certificate.not_after_time_stamp", regexValidTimeStamp),
						resource.TestMatchResourceAttr("scc_subaccount.test", "tunnel.subaccount_certificate.not_before_time_stamp", regexValidTimeStamp),
						resource.TestMatchResourceAttr("scc_subaccount.test", "tunnel.subaccount_certificate.serial_number", regexValidSerialNumber),
						resource.TestMatchResourceAttr("scc_subaccount.test", "tunnel.subaccount_certificate.subject_dn", regexp.MustCompile(`CN=.*?,L=.*?,OU=.*?,OU=.*?,O=.*?,C=.*?`)),
					),
				},
				{
					ResourceName:                         "scc_subaccount.test",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateIdFunc:                    getImportStateForSubaccount("scc_subaccount.test"),
					ImportStateVerifyIdentifierAttribute: "subaccount",
					ImportStateVerifyIgnore: []string{
						"cloud_user",
						"cloud_password",
					},
				},
				{
					ResourceName:  "scc_subaccount.test",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.com4916a705-273c-45a6-a2f0-08c234c7a23d", // malformed ID
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*subaccount.*Got:`),
				},
				{
					ResourceName:  "scc_subaccount.test",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.com,4916a705-273c-45a6-a2f0-08c234c7a23d,extra",
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
					Config: providerConfig(user) + ResourceSubaccountUpdateWithDisplayName("test", regionHost, subaccountId, user.CloudUsername, user.CloudPassword, "Initial description", "Initial Display Name"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount.test", "description", "Initial description"),
						resource.TestCheckResourceAttr("scc_subaccount.test", "display_name", "Initial Display Name"),
					),
				},
				// Update with mismatched configuration should throw error
				{
					Config:      providerConfig(user) + ResourceSubaccountUpdateWithDisplayName("test", "cf.us10.hana.ondemand.com", subaccountId, user.CloudUsername, user.CloudPassword, "Initial description", "Initial Display Name"),
					ExpectError: regexp.MustCompile(`(?is)failed to update the cloud connector subaccount due to mismatched\s+configuration values`),
				},
				{
					Config: providerConfig(user) + ResourceSubaccountUpdateWithDisplayName("test", regionHost, subaccountId, user.CloudUsername, user.CloudPassword, "Updated description", "Updated Display Name"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount.test", "description", "Updated description"),
						resource.TestCheckResourceAttr("scc_subaccount.test", "display_name", "Updated Display Name"),
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
					Config: providerConfig(user) + ResourceSubaccountWithTunnelState("test", regionHost, subaccountId, user.CloudUsername, user.CloudPassword, "Testing tunnel connected", true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount.test", "tunnel.state", "Connected"),
					),
				},
				// Update with mismatched configuration should throw error
				{
					Config:      providerConfig(user) + ResourceSubaccountWithTunnelState("test", "cf.us10.hana.ondemand.com", subaccountId, user.CloudUsername, user.CloudPassword, "Testing tunnel disconnected", false),
					ExpectError: regexp.MustCompile(`(?is)failed to update the cloud connector subaccount due to mismatched\s+configuration values`),
				},
				{
					Config: providerConfig(user) + ResourceSubaccountWithTunnelState("test", regionHost, subaccountId, user.CloudUsername, user.CloudPassword, "Testing tunnel disconnected", false),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount.test", "tunnel.state", "Disconnected"),
					),
				},
				{
					Config: providerConfig(user) + ResourceSubaccountWithTunnelState("test", regionHost, subaccountId, user.CloudUsername, user.CloudPassword, "Testing tunnel reconnected", true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount.test", "tunnel.state", "Connected"),
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
					Config:      ResourceSubaccountWoRegionHost("test", "4916a705-273c-45a6-a2f0-08c234c7a23d", user.CloudUsername, user.CloudPassword, "subaccount added via terraform tests"),
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
					Config:      ResourceSubaccountWoID("test", "cf.eu12.hana.ondemand.com", user.CloudUsername, user.CloudPassword, "subaccount added via terraform tests"),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - cloud user mandatory", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_err_wo_cloud_user")

		if len(user.CloudUsername) == 0 || len(user.CloudPassword) == 0 {
			t.Fatalf("Missing TF_VAR_cloud_user or TF_VAR_cloud_password for recording test fixtures")
		}
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubaccountWoUsername("test", "cf.eu12.hana.ondemand.com", "4916a705-273c-45a6-a2f0-08c234c7a23d", user.CloudPassword, "subaccount added via terraform tests"),
					ExpectError: regexp.MustCompile(`The argument "cloud_user" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - cloud password mandatory", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_err_wo_password")

		if len(user.CloudUsername) == 0 || len(user.CloudPassword) == 0 {
			t.Fatalf("Missing TF_VAR_cloud_user or TF_VAR_cloud_password for recording test fixtures")
		}
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubaccountWoPassword("test", "cf.eu12.hana.ondemand.com", "4916a705-273c-45a6-a2f0-08c234c7a23d", user.CloudUsername, "subaccount added via terraform tests"),
					ExpectError: regexp.MustCompile(`The argument "cloud_password" is required, but no definition was found.`),
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

func ResourceSubaccountWoUsername(datasourceName string, regionHost string, subaccount string, cloudPassword string, description string) string {
	return fmt.Sprintf(`
	resource "scc_subaccount" "%s" {
    region_host= "%s"
    subaccount= "%s"
    cloud_password= "%s" 
    description= "%s"
	}
	`, datasourceName, regionHost, subaccount, cloudPassword, description)
}

func ResourceSubaccountWoPassword(datasourceName string, regionHost string, subaccount string, cloudUser string, description string) string {
	return fmt.Sprintf(`
	resource "scc_subaccount" "%s" {
    region_host= "%s"
    subaccount= "%s"
    cloud_user= "%s"
    description= "%s"
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
