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
					Config: providerConfig(user) + ResourceSubaccountUsingAuth("scc_sa_auth", user.CloudAuthenticationData, "subaccount added via terraform tests"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "description", "subaccount added via terraform tests"),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "location_id", ""),

						resource.TestMatchResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "tunnel.connected_since", regexValidTimeStamp),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "tunnel.connections", "0"),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "tunnel.state", "Connected"),

						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "tunnel.application_connections.#", "0"),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "tunnel.service_channels.#", "0"),

						resource.TestMatchResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "tunnel.subaccount_certificate.issuer", regexp.MustCompile(`CN=.*?,OU=.*?,O=.*?,L=.*?,C=.*?`)),
						resource.TestMatchResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "tunnel.subaccount_certificate.valid_to", regexValidTimeStamp),
						resource.TestMatchResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "tunnel.subaccount_certificate.valid_from", regexValidTimeStamp),
						resource.TestMatchResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "tunnel.subaccount_certificate.serial_number", regexValidSerialNumber),
						resource.TestMatchResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "tunnel.subaccount_certificate.subject_dn", regexp.MustCompile(`CN=.*?,L=.*?,OU=.*?,OU=.*?,O=.*?,C=.*?`)),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subaccount_using_auth.scc_sa_auth",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					ResourceName:                         "scc_subaccount_using_auth.scc_sa_auth",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateIdFunc:                    getImportStateForSubaccountUsingAuth("scc_subaccount_using_auth.scc_sa_auth"),
					ImportStateVerifyIdentifierAttribute: "subaccount",
					ImportStateVerifyIgnore: []string{
						"authentication_data",
						"connected",
						"auto_renew_before_days",
					},
				},
				{
					ResourceName:  "scc_subaccount_using_auth.scc_sa_auth",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.comb1799d1c-ce91-4cd4-8b6e-dc8f4eaf0ad9", // malformed ID
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*subaccount.*Got:`),
				},
				{
					ResourceName:  "scc_subaccount_using_auth.scc_sa_auth",
					ImportState:   true,
					ImportStateId: "cf.eu12.hana.ondemand.com,b1799d1c-ce91-4cd4-8b6e-dc8f4eaf0ad9,extra",
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
					Config: providerConfig(user) + ResourceSubaccountUsingAuthUpdateWithDisplayName("scc_sa_auth", user.CloudAuthenticationData, "Initial description", "Initial Display Name"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "description", "Initial description"),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "display_name", "Initial Display Name"),
					),
				},
				{
					Config: providerConfig(user) + ResourceSubaccountUsingAuthUpdateWithDisplayName("scc_sa_auth", user.CloudAuthenticationData, "Updated description", "Updated Display Name"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "description", "Updated description"),
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "display_name", "Updated Display Name"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subaccount_using_auth.scc_sa_auth",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
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
					Config: providerConfig(user) + ResourceSubaccountUsingAuthWithTunnelState("scc_sa_auth", user.CloudAuthenticationData, "Testing tunnel connected", true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "tunnel.state", "Connected"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subaccount_using_auth.scc_sa_auth",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					Config: providerConfig(user) + ResourceSubaccountUsingAuthWithTunnelState("scc_sa_auth", user.CloudAuthenticationData, "Testing tunnel disconnected", false),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "tunnel.state", "Disconnected"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subaccount_using_auth.scc_sa_auth",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					Config: providerConfig(user) + ResourceSubaccountUsingAuthWithTunnelState("scc_sa_auth", user.CloudAuthenticationData, "Testing tunnel reconnected", true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_using_auth.scc_sa_auth", "tunnel.state", "Connected"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subaccount_using_auth.scc_sa_auth",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
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
