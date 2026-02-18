package apiobjects

type SystemCertificate struct {
	SubjectDN          string `json:"subjectDN"`
	Issuer             string `json:"issuer"`
	NotBeforeTimeStamp int64  `json:"notBeforeTimeStamp"`
	NotAfterTimeStamp  int64  `json:"notAfterTimeStamp"`
	SerialNumber       string `json:"serialNumber"`
	SubjectAltNames    string `json:"subjectAltNames,omitempty"`
}
