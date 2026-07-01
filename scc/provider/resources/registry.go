package resources

import "github.com/hashicorp/terraform-plugin-framework/resource"

func All() []func() resource.Resource {
	return []func() resource.Resource{
		NewSubaccountResource,
		NewSubaccountUsingAuthResource,
		NewSystemMappingResource,
		NewSystemMappingResourceResource,
		NewDomainMappingResource,
		NewSubaccountK8SServiceChannelResource,
		NewSubaccountABAPServiceChannelResource,
		NewSystemCertificateSelfSignedResource,
		NewSystemCertificateSignedChainResource,
		NewSystemCertificatePKCS12CertificateResource,
		NewCACertificateSelfSignedResource,
		NewCACertificateSignedChainResource,
		NewCACertificatePKCS12CertificateResource,
		NewUICertificateSelfSignedResource,
		NewUICertificateSignedChainResource,
		NewUICertificatePKCS12CertificateResource,
		NewProxySettingsResource,
		NewBackendTrustStoreResource,
	}
}
