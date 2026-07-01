package endpoints

func GetBackendTrustStoreBaseEndpoint() string {
	return "/api/v1/configuration/connector/onPremise/truststore" // The path component onPremises of the URI is available as of version 2.18.0. Older versions must use onPremise. The latter is currently accepted by all versions, but we recommend that you use onPremises as onPremise may be discontinued at some point.
}

func GetBackendTrustStoreCertificateEndpoint() string {
	return GetBackendTrustStoreBaseEndpoint() + "/certificates"
}
