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

func TestResourceSubaccountABAPServiceChannel(t *testing.T) {

	regionHost := "cf.eu12.hana.ondemand.com"
	subaccount := "d3bbbcd7-d5e0-483b-a524-6dee7205f8e8"
	abapCloudTenantHost := "testserviceid.abap.region.hana.ondemand.com"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_abap_service_channel")
		if len(user.ABAPCloudTenantHost) == 0{
			user.ABAPCloudTenantHost = abapCloudTenantHost
		}
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccountABAPServiceChannel("test", regionHost, subaccount, user.ABAPCloudTenantHost, 20, 1, false, "Created"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "region_host", regionHost),
						resource.TestMatchResourceAttr("scc_subaccount_abap_service_channel.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "abap_cloud_tenant_host", user.ABAPCloudTenantHost),
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "instance_number", "20"),
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "port", "3320"),
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "connections", "1"),
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "type", "ABAPCloud"),
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "enabled", "false"),
						resource.TestCheckResourceAttrSet("scc_subaccount_abap_service_channel.test", "id"),
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "state.connected", "false"),
						resource.TestMatchResourceAttr("scc_subaccount_abap_service_channel.test", "state.connected_since_time_stamp", regexp.MustCompile(`^(0|\d{13})$`)),
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "state.opened_connections", "0"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subaccount_abap_service_channel.test",
							map[string]knownvalue.Check{
								"id": knownvalue.Int64Func(func(v int64) error {
									if v <= 0 {
										return fmt.Errorf("id should be positive, got %d", v)
									}
									return nil
								}),
								"region_host": knownvalue.StringExact(regionHost),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					ResourceName:      "scc_subaccount_abap_service_channel.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: getImportStateForSubaccountABAPServiceChannel("scc_subaccount_abap_service_channel.test"),
				},
				{
					ResourceName:  "scc_subaccount_abap_service_channel.test",
					ImportState:   true,
					ImportStateId: regionHost + subaccount + "1", // malformed ID
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*id.*Got:`),
				},
				{
					ResourceName:  "scc_subaccount_abap_service_channel.test",
					ImportState:   true,
					ImportStateId: regionHost + "," + subaccount + ",1" + ",extra",
					ExpectError:   regexp.MustCompile(`(?is)Expected import identifier with format:.*id.*Got:`),
				},
				{
					ResourceName:  "scc_subaccount_abap_service_channel.test",
					ImportState:   true,
					ImportStateId: regionHost + "," + subaccount + ",not-an-int",
					ExpectError:   regexp.MustCompile(`(?is)The 'id' part must be an integer.*Got:.*not-an-int`),
				},
			},
		})

	})

	t.Run("update path - comment and connections update", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount_abap_service_channel_update")
		if len(user.ABAPCloudTenantHost) == 0{
			user.ABAPCloudTenantHost = abapCloudTenantHost
		}
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccountABAPServiceChannel("test", regionHost, subaccount, user.ABAPCloudTenantHost, 20, 1, false, "Created"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "comment", "Created"),
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "connections", "1"),
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "enabled", "false"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subaccount_abap_service_channel.test",
							map[string]knownvalue.Check{
								"id": knownvalue.Int64Func(func(v int64) error {
									if v <= 0 {
										return fmt.Errorf("id should be positive, got %d", v)
									}
									return nil
								}),
								"region_host": knownvalue.StringExact(regionHost),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				// Update with mismatched configuration should throw error
				{
					Config:      providerConfig(user) + ResourceSubaccountABAPServiceChannel("test", "cf.us10.hana.ondemand.com", subaccount, user.ABAPCloudTenantHost, 20, 1, false, "Update"),
					ExpectError: regexp.MustCompile(`(?is)error updating the cloud connector subaccount ABAP service channel.*mismatched\s+configuration values`),
				},
				{
					Config: providerConfig(user) + ResourceSubaccountABAPServiceChannel("test", regionHost, subaccount, user.ABAPCloudTenantHost, 20, 2, false, "Changed Connections"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "comment", "Changed Connections"),
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "connections", "2"),
						resource.TestCheckResourceAttr("scc_subaccount_abap_service_channel.test", "enabled", "false"),
					),
				},
			},
		})
	})

	t.Run("error path - region host mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubaccountABAPServiceChannelWoRegionHost("test", subaccount, abapCloudTenantHost, 20, 1),
					ExpectError: regexp.MustCompile(`(?s)The argument\s+"region_host"\s+is required, but no definition was\s+found\.`),
				},
			},
		})
	})

	t.Run("error path - subaccount id mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubaccountABAPServiceChannelWoSubaccount("test", regionHost, abapCloudTenantHost, 20, 1),
					ExpectError: regexp.MustCompile(`(?s)The argument\s+"subaccount"\s+(is required, but no definition was\s+found|value must be a valid UUID)`),
				},
			},
		})
	})

	t.Run("error path - abap cloud tenant host mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubaccountABAPServiceChannelWoABAPCloudTenantHost("test", regionHost, subaccount, 20, 1),
					ExpectError: regexp.MustCompile(`(?s)The argument\s+"abap_cloud_tenant_host"\s+is required, but no definition was\s+found\.`),
				},
			},
		})
	})

	t.Run("error path - instance_number mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubaccountABAPServiceChannelWoInstanceNumber("test", regionHost, subaccount, abapCloudTenantHost, 1),
					ExpectError: regexp.MustCompile(`(?s)The argument\s+"instance_number"\s+is required, but no definition was\s+found\.`),
				},
			},
		})
	})

	t.Run("error path - connections mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubaccountABAPServiceChannelWoConnections("test", regionHost, subaccount, abapCloudTenantHost, 20),
					ExpectError: regexp.MustCompile(`(?s)The argument\s+"connections"\s+is required, but no definition was\s+found\.`),
				},
			},
		})
	})

}

func ResourceSubaccountABAPServiceChannel(resourceName string, regionHost string, subaccount string, abapCloudTenantHost string, instanceNumber int64, connections int64, enabled bool, comment string) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_abap_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	abap_cloud_tenant_host =  "%s"
	instance_number =  "%d"
	connections = "%d"
	enabled = "%t"
	comment = "%s"
	}
	`, resourceName, regionHost, subaccount, abapCloudTenantHost, instanceNumber, connections, enabled, comment)
}

func ResourceSubaccountABAPServiceChannelWoRegionHost(resourceName string, subaccount string, abapCloudTenantHost string, instanceNumber int64, connections int64) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_abap_service_channel" "%s" {
	subaccount = "%s"
	abap_cloud_tenant_host =  "%s"
	instance_number =  "%d"
	connections = "%d"
	}
	`, resourceName, subaccount, abapCloudTenantHost, instanceNumber, connections)
}

func ResourceSubaccountABAPServiceChannelWoSubaccount(resourceName string, regionHost string, abapCloudTenantHost string, instanceNumber int64, connections int64) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_abap_service_channel" "%s" {
	region_host = "%s"
	abap_cloud_tenant_host =  "%s"
	instance_number =  "%d"
	connections = "%d"
	}
	`, resourceName, regionHost, abapCloudTenantHost, instanceNumber, connections)
}

func ResourceSubaccountABAPServiceChannelWoABAPCloudTenantHost(resourceName string, regionHost string, subaccount string, instanceNumber int64, connections int64) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_abap_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	instance_number =  "%d"
	connections = "%d"
	}
	`, resourceName, regionHost, subaccount, instanceNumber, connections)
}

func ResourceSubaccountABAPServiceChannelWoInstanceNumber(resourceName string, regionHost string, subaccount string, abapCloudTenantHost string, connections int64) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_abap_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	abap_cloud_tenant_host =  "%s"
	connections = "%d"
	}
	`, resourceName, regionHost, subaccount, abapCloudTenantHost, connections)
}

func ResourceSubaccountABAPServiceChannelWoConnections(resourceName string, regionHost string, subaccount string, abapCloudTenantHost string, instanceNumber int64) string {
	return fmt.Sprintf(`
	resource "scc_subaccount_abap_service_channel" "%s" {
	region_host = "%s"
	subaccount = "%s"
	abap_cloud_tenant_host =  "%s"
	instance_number =  "%d"
	}
	`, resourceName, regionHost, subaccount, abapCloudTenantHost, instanceNumber)
}

func getImportStateForSubaccountABAPServiceChannel(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s,%s",
			rs.Primary.Attributes["region_host"],
			rs.Primary.Attributes["subaccount"],
			rs.Primary.Attributes["id"],
		), nil
	}
}
