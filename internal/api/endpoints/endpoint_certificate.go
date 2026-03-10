package endpoints

func GetCACertificateEndpoint() string {
	return "/api/v1/configuration/connector/onPremise/ppCaCertificate"
}

func GetSystemCertificateEndpoint() string {
	return "/api/v1/configuration/connector/onPremise/systemCertificate"
}

func GetUICertificateEndpoint() string {
	return "/api/v1/configuration/connector/ui/uiCertificate"
}
