package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceSubaccountUsingAuth(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_using_auth")
		if len(user.CloudAuthenticationData) == 0 {
			t.Fatalf("Missing TF_VAR_authentication_data for recording test fixtures")
		}
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccountUsingAuth("test", user.CloudAuthenticationData, "subaccount added via terraform tests"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("scc_subaccount_using_auth.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "description", "subaccount added via terraform tests"),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "location_id", ""),

						resource.TestMatchResourceAttr("scc_subaccount_using_auth.test", "tunnel.connected_since_time_stamp", regexValidTimeStamp),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "tunnel.connections", "0"),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "tunnel.state", "Connected"),

						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "tunnel.application_connections.#", "0"),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "tunnel.service_channels.#", "0"),

						resource.TestMatchResourceAttr("scc_subaccount_using_auth.test", "tunnel.subaccount_certificate.issuer", regexp.MustCompile(`CN=.*?,OU=S.*?,O=.*?,L=.*?,C=.*?`)),
						resource.TestMatchResourceAttr("scc_subaccount_using_auth.test", "tunnel.subaccount_certificate.not_after_time_stamp", regexValidTimeStamp),
						resource.TestMatchResourceAttr("scc_subaccount_using_auth.test", "tunnel.subaccount_certificate.not_before_time_stamp", regexValidTimeStamp),
						resource.TestMatchResourceAttr("scc_subaccount_using_auth.test", "tunnel.subaccount_certificate.serial_number", regexValidSerialNumber),
						resource.TestMatchResourceAttr("scc_subaccount_using_auth.test", "tunnel.subaccount_certificate.subject_dn", regexp.MustCompile(`CN=.*?,L=.*?,OU=.*?,OU=.*?,O=.*?,C=.*?`)),
					),
				},
				{
					ResourceName:                         "scc_subaccount_using_auth.test",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateIdFunc:                    getImportStateForSubaccountUsingAuth("scc_subaccount_using_auth.test"),
					ImportStateVerifyIdentifierAttribute: "subaccount",
					ImportStateVerifyIgnore: []string{
						"authentication_data",
						"connected",
					},
				},
				{
					ResourceName:  "scc_subaccount_using_auth.test",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.com4916a705-273c-45a6-a2f0-08c234c7a23d", // malformed ID
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*subaccount.*Got:`),
				},
				{
					ResourceName:  "scc_subaccount_using_auth.test",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.com,4916a705-273c-45a6-a2f0-08c234c7a23d,extra",
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*subaccount.*Got:`),
				},
			},
		})

	})

	t.Run("update path - update description and display name", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_using_auth_update")
		if len(user.CloudUsername) == 0 || len(user.CloudPassword) == 0 {
			t.Fatalf("Missing TF_VAR_cloud_user or TF_VAR_cloud_password for recording test fixtures")
		}
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccountUsingAuthUpdateWithDisplayName("test", user.CloudAuthenticationData, "Initial description", "Initial Display Name"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "description", "Initial description"),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "display_name", "Initial Display Name"),
					),
				},
				{
					Config: providerConfig(user) + ResourceSubaccountUsingAuthUpdateWithDisplayName("test", user.CloudAuthenticationData, "Updated description", "Updated Display Name"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "description", "Updated description"),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "display_name", "Updated Display Name"),
					),
				},
			},
		})
	})

	t.Run("update path - tunnel state change", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_using_auth_update_tunnel")
		if user.CloudUsername == "" || user.CloudPassword == "" {
			t.Fatalf("Missing TF_VAR_cloud_user or TF_VAR_cloud_password for recording test fixtures")
		}
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccountUsingAuthWithTunnelState("test", user.CloudAuthenticationData, "Testing tunnel connected", true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "tunnel.state", "Connected"),
					),
				},
				{
					Config: providerConfig(user) + ResourceSubaccountUsingAuthWithTunnelState("test", user.CloudAuthenticationData, "Testing tunnel disconnected", false),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "tunnel.state", "Disconnected"),
					),
				},
				{
					Config: providerConfig(user) + ResourceSubaccountUsingAuthWithTunnelState("test", user.CloudAuthenticationData, "Testing tunnel reconnected", true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.test", "tunnel.state", "Connected"),
					),
				},
			},
		})
	})

	t.Run("error path - authentication data mandatory", func(t *testing.T) {
		rec, _ := setupVCR(t, "fixtures/resource_subaccount_using_auth_err_wo_authentication_data")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubaccountUsingAuthWoAuthenticationData("test", "subaccount added via terraform tests"),
					ExpectError: regexp.MustCompile(`The argument "authentication_data" is required, but no definition was found.`),
				},
			},
		})
	})
}

func ResourceSubaccountUsingAuth(datasourceName, authenticationData, description string) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_using_auth" "%s" {
    authentication_data = "%s"
    description= "%s"
	}
	`, datasourceName, authenticationData, description)
}

func ResourceSubaccountUsingAuthWoAuthenticationData(datasourceName, description string) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_using_auth" "%s" {
    description= "%s"
	}
	`, datasourceName, description)
}

func ResourceSubaccountUsingAuthUpdateWithDisplayName(datasourceName, authenticationData, description, displayName string) string {
	return fmt.Sprintf(`
resource "scc_subaccount_using_auth" "%s" {
  authentication_data = "%s"
  description   = "%s"
  display_name  = "%s"
}
`, datasourceName, authenticationData, description, displayName)
}

func ResourceSubaccountUsingAuthWithTunnelState(datasourceName, authenticationData, description string, connected bool) string {
	return fmt.Sprintf(`
resource "scc_subaccount_using_auth" "%s" {
  authentication_data = "%s"
  description    = "%s"
  connected = %t
}
`, datasourceName, authenticationData, description, connected)
}

func getImportStateForSubaccountUsingAuth(resourceName string) resource.ImportStateIdFunc {
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
