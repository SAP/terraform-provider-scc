package provider

import (
	"context"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

type testResource struct {
	name      string
	resource  resource.ResourceWithConfigure
	getClient func(resource.Resource) *api.RestApiClient
}

var resources = []testResource{
	{
		name:     "SubaccountResource",
		resource: &SubaccountResource{},
		getClient: func(r resource.Resource) *api.RestApiClient {
			return r.(*SubaccountResource).client
		},
	},
	{
		name:     "SystemMappingResource",
		resource: &SystemMappingResource{},
		getClient: func(r resource.Resource) *api.RestApiClient {
			return r.(*SystemMappingResource).client
		},
	},
	{
		name:     "SystemMappingResourceResource",
		resource: &SystemMappingResourceResource{},
		getClient: func(r resource.Resource) *api.RestApiClient {
			return r.(*SystemMappingResourceResource).client
		},
	},
	{
		name:     "DomainMappingResource",
		resource: &DomainMappingResource{},
		getClient: func(r resource.Resource) *api.RestApiClient {
			return r.(*DomainMappingResource).client
		},
	},
	{
		name:     "SubaccountK8SServiceChannelResource",
		resource: &SubaccountK8SServiceChannelResource{},
		getClient: func(r resource.Resource) *api.RestApiClient {
			return r.(*SubaccountK8SServiceChannelResource).client
		},
	},
}

func TestAllResourceConfigure(t *testing.T) {
	mockClient := &api.RestApiClient{}

	for _, tr := range resources {
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
