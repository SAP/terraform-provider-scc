package apiobjects

type BackendTrustStoreConfiguration struct {
	TrustAllBackends bool              `json:"trustAllBackends"`
	TrustedBackends  []TrustedBackends `json:"trustedBackends"`
}

type TrustedBackends struct {
	Alias     string `json:"alias"`
	SubjectDN string `json:"subjectDN"`
	Issuer    string `json:"issuer"`
	ValidTo   int64  `json:"notAfterTimeStamp"`
}
