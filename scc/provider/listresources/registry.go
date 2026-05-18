package listresources

import "github.com/hashicorp/terraform-plugin-framework/list"

func All() []func() list.ListResource {
	return []func() list.ListResource{
		NewSubaccountListResource,
		NewDomainMappingListResource,
		NewSystemMappingListResource,
		NewSystemMappingResourceListResource,
		NewSubaccountABAPServiceChannelListResource,
		NewSubaccountK8SServiceChannelListResource,
	}
}
