package resources_test

import (
	"context"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/SAP/terraform-provider-scc/scc/provider/resources"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

type testResource struct {
	name      string
	resource  resource.ResourceWithConfigure
	getClient func(resource.Resource) *api.RestApiClient
}

var all_resources = []testResource{
	{
		name:     "SubaccountResource",
		resource: &resources.SubaccountResource{},
		getClient: func(r resource.Resource) *api.RestApiClient {
			return r.(*resources.SubaccountResource).Client
		},
	},
	{
		name:     "SystemMappingResource",
		resource: &resources.SystemMappingResource{},
		getClient: func(r resource.Resource) *api.RestApiClient {
			return r.(*resources.SystemMappingResource).Client
		},
	},
	{
		name:     "SystemMappingResourceResource",
		resource: &resources.SystemMappingResourceResource{},
		getClient: func(r resource.Resource) *api.RestApiClient {
			return r.(*resources.SystemMappingResourceResource).Client
		},
	},
	{
		name:     "DomainMappingResource",
		resource: &resources.DomainMappingResource{},
		getClient: func(r resource.Resource) *api.RestApiClient {
			return r.(*resources.DomainMappingResource).Client
		},
	},
	{
		name:     "SubaccountK8SServiceChannelResource",
		resource: &resources.SubaccountK8SServiceChannelResource{},
		getClient: func(r resource.Resource) *api.RestApiClient {
			return r.(*resources.SubaccountK8SServiceChannelResource).Client
		},
	},
}

func TestAllResourceConfigure(t *testing.T) {
	mockClient := &api.RestApiClient{}

	for _, tr := range all_resources {
		t.Run(tr.name+"_nil_provider_data", func(t *testing.T) {
			resp := &resource.ConfigureResponse{}
			tr.resource.Configure(context.Background(), resource.ConfigureRequest{ProviderData: nil}, resp)

			assert.Nil(t, tr.getClient(tr.resource), "Expected nil client for nil ProviderData")
			assert.False(t, resp.Diagnostics.HasError(), "Expected no error for nil ProviderData")
		})

		t.Run(tr.name+"_invalid_provider_data", func(t *testing.T) {
			resp := &resource.ConfigureResponse{}
			tr.resource.Configure(context.Background(), resource.ConfigureRequest{ProviderData: "invalid-type"}, resp)

			assert.Nil(t, tr.getClient(tr.resource), "Expected nil client for invalid ProviderData")
			assert.True(t, resp.Diagnostics.HasError(), "Expected error for invalid ProviderData")
		})

		t.Run(tr.name+"_valid_provider_data", func(t *testing.T) {
			resp := &resource.ConfigureResponse{}
			tr.resource.Configure(context.Background(), resource.ConfigureRequest{ProviderData: mockClient}, resp)

			assert.Equal(t, mockClient, tr.getClient(tr.resource), "Expected client to be set")
			assert.False(t, resp.Diagnostics.HasError(), "Expected no error for valid ProviderData")
		})
	}
}
