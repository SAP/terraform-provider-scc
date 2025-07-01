package provider

import (
	"context"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
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
		datasource: &SubaccountConfigurationDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*SubaccountConfigurationDataSource).client
		},
	},
	{
		name:       "SubaccountsDataSource",
		datasource: &SubaccountsDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*SubaccountsDataSource).client
		},
	},
	{
		name:       "SystemMappingDataSource",
		datasource: &SystemMappingDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*SystemMappingDataSource).client
		},
	},
	{
		name:       "SystemMappingsDataSource",
		datasource: &SystemMappingsDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*SystemMappingsDataSource).client
		},
	},
	{
		name:       "SystemMappingResourceDataSource",
		datasource: &SystemMappingResourceDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*SystemMappingResourceDataSource).client
		},
	},
	{
		name:       "SystemMappingResourcesDataSource",
		datasource: &SystemMappingResourcesDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*SystemMappingResourcesDataSource).client
		},
	},
	{
		name:       "DomainMappingDataSource",
		datasource: &DomainMappingDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*DomainMappingDataSource).client
		},
	},
	{
		name:       "DomainMappingsDataSource",
		datasource: &DomainMappingsDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*DomainMappingsDataSource).client
		},
	},
	{
		name:       "SubaccountK8SServiceChannelDataSource",
		datasource: &SubaccountK8SServiceChannelDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*SubaccountK8SServiceChannelDataSource).client
		},
	},
	{
		name:       "SubaccountK8SServiceChannelsDataSource",
		datasource: &SubaccountK8SServiceChannelsDataSource{},
		getClient: func(r datasource.DataSource) *api.RestApiClient {
			return r.(*SubaccountK8SServiceChannelsDataSource).client
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
