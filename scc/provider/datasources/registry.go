package datasources

import "github.com/hashicorp/terraform-plugin-framework/datasource"

func All() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSubaccountsDataSource,
		NewSubaccountConfigurationDataSource,
		NewSystemMappingsDataSource,
		NewSystemMappingDataSource,
		NewSystemMappingResourcesDataSource,
		NewSystemMappingResourceDataSource,
		NewDomainMappingsDataSource,
		NewDomainMappingDataSource,
		NewSubaccountK8SServiceChannelDataSource,
		NewSubaccountK8SServiceChannelsDataSource,
		NewSubaccountABAPServiceChannelDataSource,
		NewSubaccountABAPServiceChannelsDataSource,
		NewSystemCertificateDataSource,
		NewCACertificateDataSource,
		NewProxySettingsDataSource,
		NewBackendTrustStoreDataSource,
	}
}
