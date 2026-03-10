package apiobjects

type Certificate struct {
	SubjectDN          string            `json:"subjectDN"`
	Issuer             string            `json:"issuer"`
	NotBeforeTimeStamp int64             `json:"notBeforeTimeStamp"`
	NotAfterTimeStamp  int64             `json:"notAfterTimeStamp"`
	SerialNumber       string            `json:"serialNumber"`
	SubjectAltNames    []SubjectAltNames `json:"subjectAltNames,omitempty"`
}

type SubjectAltNames struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
