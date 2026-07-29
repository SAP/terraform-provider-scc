package apiobjects

type SubjectPatternRules []SubjectPatternRule

type SubjectPatternRule struct {
	Description    string         `json:"description"`
	Condition      string         `json:"condition"`
	SubjectPattern SubjectPattern `json:"subjectPattern"`
}

type SubjectPattern struct {
	CommonName       string `json:"CN,omitempty"`
	Email            string `json:"EMAIL,omitempty"`
	Locality         string `json:"L,omitempty"`
	OrganizationUnit string `json:"OU,omitempty"`
	Organization     string `json:"O,omitempty"`
	State            string `json:"ST,omitempty"`
	Country          string `json:"C,omitempty"`
}
