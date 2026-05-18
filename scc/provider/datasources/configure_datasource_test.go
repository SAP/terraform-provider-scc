package datasources_test

import (
	"context"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/SAP/terraform-provider-scc/scc/provider/datasources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

type testDataSource struct {
	name       string
	datasource datasource.DataSourceWithConfigure
	getClient  func(datasource.DataSource) *api.RestApiClient
}

var dataSources = []testDataSource{
	{
		name:       "SubaccountDataSource",
		datasource: &datasources.SubaccountConfigurationDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*datasources.SubaccountConfigurationDataSource).Client
		},
	},
	{
		name:       "SubaccountsDataSource",
		datasource: &datasources.SubaccountsDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*datasources.SubaccountsDataSource).Client
		},
	},
	{
		name:       "SystemMappingDataSource",
		datasource: &datasources.SystemMappingDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*datasources.SystemMappingDataSource).Client
		},
	},
	{
		name:       "SystemMappingsDataSource",
		datasource: &datasources.SystemMappingsDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*datasources.SystemMappingsDataSource).Client
		},
	},
	{
		name:       "SystemMappingResourceDataSource",
		datasource: &datasources.SystemMappingResourceDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*datasources.SystemMappingResourceDataSource).Client
		},
	},
	{
		name:       "SystemMappingResourcesDataSource",
		datasource: &datasources.SystemMappingResourcesDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*datasources.SystemMappingResourcesDataSource).Client
		},
	},
	{
		name:       "DomainMappingDataSource",
		datasource: &datasources.DomainMappingDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*datasources.DomainMappingDataSource).Client
		},
	},
	{
		name:       "DomainMappingsDataSource",
		datasource: &datasources.DomainMappingsDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*datasources.DomainMappingsDataSource).Client
		},
	},
	{
		name:       "SubaccountK8SServiceChannelDataSource",
		datasource: &datasources.SubaccountK8SServiceChannelDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*datasources.SubaccountK8SServiceChannelDataSource).Client
		},
	},
	{
		name:       "SubaccountK8SServiceChannelsDataSource",
		datasource: &datasources.SubaccountK8SServiceChannelsDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*datasources.SubaccountK8SServiceChannelsDataSource).Client
		},
	},
}

func TestAllDataSourceConfigure(t *testing.T) {
	mockClient := &api.RestApiClient{}

	for _, td := range dataSources {
		t.Run(td.name+"_nil_provider_data", func(t *testing.T) {
			resp := &datasource.ConfigureResponse{}
			td.datasource.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: nil}, resp)

			assert.Nil(t, td.getClient(td.datasource), "Expected nil client for nil ProviderData")
			assert.False(t, resp.Diagnostics.HasError(), "Expected no error for nil ProviderData")
		})

		t.Run(td.name+"_invalid_provider_data", func(t *testing.T) {
			resp := &datasource.ConfigureResponse{}
			td.datasource.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "invalid-type"}, resp)

			assert.Nil(t, td.getClient(td.datasource), "Expected nil client for invalid ProviderData")
			assert.True(t, resp.Diagnostics.HasError(), "Expected error for invalid ProviderData")
		})

		t.Run(td.name+"_valid_provider_data", func(t *testing.T) {
			resp := &datasource.ConfigureResponse{}
			td.datasource.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, resp)

			assert.Equal(t, mockClient, td.getClient(td.datasource), "Expected client to be set")
			assert.False(t, resp.Diagnostics.HasError(), "Expected no error for valid ProviderData")
		})
	}
}
