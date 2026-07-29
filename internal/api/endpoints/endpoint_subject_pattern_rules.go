package endpoints

import "fmt"

func GetSubjectPatternRulesBaseEndpoint() string {
	return "/api/v1/configuration/connector/onPremises/subjectPatternRules"
}

func GetSubjectPatternRuleByIndexEndpoint(index int) string {
	return GetSubjectPatternRulesBaseEndpoint() + "/" + fmt.Sprint(index)
}
