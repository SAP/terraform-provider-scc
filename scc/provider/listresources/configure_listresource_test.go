package listresources_test

import (
	"context"
	"testing"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/SAP/terraform-provider-scc/scc/provider/listresources"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

type testListResource struct {
	name         string
	listresource list.ListResourceWithConfigure
	getClient    func(list.ListResource) *api.RestApiClient
}

var listResources = []testListResource{
	{
		name:         "DomainMappingListResource",
		listresource: &listresources.DomainMappingListResource{},
		getClient: func(r list.ListResource) *api.RestApiClient {
			return r.(*listresources.DomainMappingListResource).Client
		},
	},
	{
		name:         "SubaccountListResource",
		listresource: &listresources.SubaccountListResource{},
		getClient: func(r list.ListResource) *api.RestApiClient {
			return r.(*listresources.SubaccountListResource).Client
		},
	},
	{
		name:         "SubaccountABAPServiceChannelListResource",
		listresource: &listresources.SubaccountABAPServiceChannelListResource{},
		getClient: func(r list.ListResource) *api.RestApiClient {
			return r.(*listresources.SubaccountABAPServiceChannelListResource).Client
		},
	},
	{
		name:         "SubaccountK8SServiceChannelListResource",
		listresource: &listresources.SubaccountK8SServiceChannelListResource{},
		getClient: func(r list.ListResource) *api.RestApiClient {
			return r.(*listresources.SubaccountK8SServiceChannelListResource).Client
		},
	},
	{
		name:         "SubjectPatternRuleListResource",
		listresource: &listresources.SubjectPatternRuleListResource{},
		getClient: func(r list.ListResource) *api.RestApiClient {
			return r.(*listresources.SubjectPatternRuleListResource).Client
		},
	},
	{
		name:         "SystemMappingListResource",
		listresource: &listresources.SystemMappingListResource{},
		getClient: func(r list.ListResource) *api.RestApiClient {
			return r.(*listresources.SystemMappingListResource).Client
		},
	},
	{
		name:         "SystemMappingResourceListResource",
		listresource: &listresources.SystemMappingResourceListResource{},
		getClient: func(r list.ListResource) *api.RestApiClient {
			return r.(*listresources.SystemMappingResourceListResource).Client
		},
	},
}

func TestAllListResourceConfigure(t *testing.T) {
	mockClient := &api.RestApiClient{}

	for _, tlr := range listResources {
		t.Run(tlr.name+"_nil_provider_data", func(t *testing.T) {
			resp := &resource.ConfigureResponse{}
			tlr.listresource.Configure(context.Background(), resource.ConfigureRequest{ProviderData: nil}, resp)

			assert.Nil(t, tlr.getClient(tlr.listresource), "Expected nil client for nil ProviderData")
			assert.False(t, resp.Diagnostics.HasError(), "Expected no error for nil ProviderData")
		})

		t.Run(tlr.name+"_invalid_provider_data", func(t *testing.T) {
			resp := &resource.ConfigureResponse{}
			tlr.listresource.Configure(context.Background(), resource.ConfigureRequest{ProviderData: "invalid-type"}, resp)

			assert.Nil(t, tlr.getClient(tlr.listresource), "Expected nil client for invalid ProviderData")
			assert.True(t, resp.Diagnostics.HasError(), "Expected error for invalid ProviderData")
		})

		t.Run(tlr.name+"_valid_provider_data", func(t *testing.T) {
			resp := &resource.ConfigureResponse{}
			tlr.listresource.Configure(context.Background(), resource.ConfigureRequest{ProviderData: mockClient}, resp)

			assert.Equal(t, mockClient, tlr.getClient(tlr.listresource), "Expected client to be set")
			assert.False(t, resp.Diagnostics.HasError(), "Expected no error for valid ProviderData")
		})
	}
}
