package resources_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type TestSubjectPatternRuleCondition struct {
	Variable string
	Operator string
	Value    string
}

type TestSubjectPattern struct {
	CommonName       string
	Email            string
	Locality         string
	OrganizationUnit string
	Organization     string
	State            string
	Country          string
}

// generateCondition creates a TestSubjectPatternRuleCondition as an HCL object literal.
func generateCondition(c TestSubjectPatternRuleCondition) string {
	var b strings.Builder
	b.WriteString("{\n")
	if c.Variable != "" {
		fmt.Fprintf(&b, "    variable = %q\n", c.Variable)
	}
	if c.Operator != "" {
		fmt.Fprintf(&b, "    operator = %q\n", c.Operator)
	}
	if c.Value != "" {
		fmt.Fprintf(&b, "    value    = %q\n", c.Value)
	}
	b.WriteString("  }")
	return b.String()
}

// generateSubjectPattern creates a TestSubjectPattern as an HCL object literal.
func generateSubjectPattern(sp TestSubjectPattern) string {
	var b strings.Builder
	b.WriteString("{\n")
	if sp.CommonName != "" {
		fmt.Fprintf(&b, "    cn    = %q\n", sp.CommonName)
	}
	if sp.Email != "" {
		fmt.Fprintf(&b, "    email = %q\n", sp.Email)
	}
	if sp.Locality != "" {
		fmt.Fprintf(&b, "    l     = %q\n", sp.Locality)
	}
	if sp.OrganizationUnit != "" {
		fmt.Fprintf(&b, "    ou    = %q\n", sp.OrganizationUnit)
	}
	if sp.Organization != "" {
		fmt.Fprintf(&b, "    o     = %q\n", sp.Organization)
	}
	if sp.State != "" {
		fmt.Fprintf(&b, "    st    = %q\n", sp.State)
	}
	if sp.Country != "" {
		fmt.Fprintf(&b, "    c     = %q\n", sp.Country)
	}
	b.WriteString("  }")
	return b.String()
}

var defaultCondition = TestSubjectPatternRuleCondition{
	Variable: "name",
	Operator: "is",
	Value:    "Terraform",
}

var defaultSubjectPattern = TestSubjectPattern{
	CommonName:       "Terraform",
	Email:            "terraform@example.com",
	Locality:         "Test Locality",
	OrganizationUnit: "Test OU",
	Organization:     "Test Org",
	State:            "Test State",
	Country:          "TC",
}

func TestResourceSubjectPatternRule(t *testing.T) {
	t.Parallel()

	t.Run("happy path: subject pattern rule full config", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/resource_subject_pattern_rule")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: tfutils.ProviderConfig(user) + ResourceSubjectPatternRule("scc_spr", "subject pattern rule added via terraform tests", defaultCondition, defaultSubjectPattern),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("scc_subject_pattern_rule.scc_spr", "index"),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "description", "subject pattern rule added via terraform tests"),

						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "condition.variable", defaultCondition.Variable),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "condition.operator", defaultCondition.Operator),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "condition.value", defaultCondition.Value),

						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.cn", defaultSubjectPattern.CommonName),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.email", defaultSubjectPattern.Email),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.l", defaultSubjectPattern.Locality),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.ou", defaultSubjectPattern.OrganizationUnit),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.o", defaultSubjectPattern.Organization),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.st", defaultSubjectPattern.State),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.c", defaultSubjectPattern.Country),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subject_pattern_rule.scc_spr",
							map[string]knownvalue.Check{
								"index": knownvalue.NotNull(),
							},
						),
					},
				},
				{
					ResourceName:                         "scc_subject_pattern_rule.scc_spr",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateIdFunc:                    getImportStateForSubjectPatternRule("scc_subject_pattern_rule.scc_spr"),
					ImportStateVerifyIdentifierAttribute: "index",
				},
			},
		})
	})

	t.Run("update path - update description, condition and subject pattern", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/resource_subject_pattern_rule.update")
		defer tfutils.StopQuietly(rec)

		updatedCondition := TestSubjectPatternRuleCondition{
			Variable: "login_name",
			Operator: "is",
			Value:    "UpdatedValue",
		}
		updatedSubjectPattern := TestSubjectPattern{
			CommonName: "UpdatedCN",
		}

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: tfutils.ProviderConfig(user) + ResourceSubjectPatternRule("scc_spr", "subject pattern rule added via terraform tests", defaultCondition, defaultSubjectPattern),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("scc_subject_pattern_rule.scc_spr", "index"),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "description", "subject pattern rule added via terraform tests"),

						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "condition.variable", defaultCondition.Variable),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "condition.operator", defaultCondition.Operator),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "condition.value", defaultCondition.Value),

						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.cn", defaultSubjectPattern.CommonName),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.email", defaultSubjectPattern.Email),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.l", defaultSubjectPattern.Locality),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.ou", defaultSubjectPattern.OrganizationUnit),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.o", defaultSubjectPattern.Organization),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.st", defaultSubjectPattern.State),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.c", defaultSubjectPattern.Country),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subject_pattern_rule.scc_spr",
							map[string]knownvalue.Check{
								"index": knownvalue.NotNull(),
							},
						),
					},
				},
				{
					Config: tfutils.ProviderConfig(user) + ResourceSubjectPatternRule("scc_spr", "updated description", updatedCondition, updatedSubjectPattern),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("scc_subject_pattern_rule.scc_spr", "index"),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "description", "updated description"),

						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "condition.variable", updatedCondition.Variable),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "condition.operator", updatedCondition.Operator),
						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "condition.value", updatedCondition.Value),

						resource.TestCheckResourceAttr("scc_subject_pattern_rule.scc_spr", "subject_pattern.cn", updatedSubjectPattern.CommonName),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity(
							"scc_subject_pattern_rule.scc_spr",
							map[string]knownvalue.Check{
								"index": knownvalue.NotNull(),
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - description mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubjectPatternRuleWoDescription("scc_spr", defaultCondition, defaultSubjectPattern),
					ExpectError: regexp.MustCompile(`(?s).*"description".*required`),
				},
			},
		})
	})

	t.Run("error path - condition mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubjectPatternRuleWoCondition("scc_spr", "Test Description", defaultSubjectPattern),
					ExpectError: regexp.MustCompile(`(?s).*"condition".*required`),
				},
			},
		})
	})

	t.Run("error path - subject_pattern mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      ResourceSubjectPatternRuleWoSubjectPattern("scc_spr", "Test Description", defaultCondition),
					ExpectError: regexp.MustCompile(`(?s).*"subject_pattern".*required`),
				},
			},
		})
	})

	t.Run("error path - condition.variable mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config: ResourceSubjectPatternRule("scc_spr", "Test Description",
						TestSubjectPatternRuleCondition{Operator: "is", Value: "Terraform"},
						defaultSubjectPattern,
					),
					ExpectError: regexp.MustCompile(`(?s).*"variable".*required`),
				},
			},
		})
	})

	t.Run("error path - condition.operator mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config: ResourceSubjectPatternRule("scc_spr", "Test Description",
						TestSubjectPatternRuleCondition{Variable: "name", Value: "Terraform"},
						defaultSubjectPattern,
					),
					ExpectError: regexp.MustCompile(`(?s).*"operator".*required`),
				},
			},
		})
	})

	t.Run("error path - condition.operator invalid value", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config: ResourceSubjectPatternRule("scc_spr", "Test Description",
						TestSubjectPatternRuleCondition{Variable: "name", Operator: "equals", Value: "Terraform"},
						defaultSubjectPattern,
					),
					ExpectError: regexp.MustCompile(`(?s).*condition\.operator.*one of`),
				},
			},
		})
	})

	t.Run("error path - user_type with exist operator", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config: ResourceSubjectPatternRule("scc_spr", "Test Description",
						TestSubjectPatternRuleCondition{Variable: "user_type", Operator: "exist"},
						defaultSubjectPattern,
					),
					ExpectError: regexp.MustCompile(`(?s).*Invalid operator`),
				},
			},
		})
	})

	t.Run("error path - user_type with exist_not operator", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config: ResourceSubjectPatternRule("scc_spr", "Test Description",
						TestSubjectPatternRuleCondition{Variable: "user_type", Operator: "exist_not"},
						defaultSubjectPattern,
					),
					ExpectError: regexp.MustCompile(`(?s).*Invalid operator`),
				},
			},
		})
	})

	t.Run("error path - is operator without value", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config: ResourceSubjectPatternRule("scc_spr", "Test Description",
						TestSubjectPatternRuleCondition{Variable: "name", Operator: "is"},
						defaultSubjectPattern,
					),
					ExpectError: regexp.MustCompile(`(?s).*Missing value`),
				},
			},
		})
	})

	t.Run("error path - is_not operator without value", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config: ResourceSubjectPatternRule("scc_spr", "Test Description",
						TestSubjectPatternRuleCondition{Variable: "name", Operator: "is_not"},
						defaultSubjectPattern,
					),
					ExpectError: regexp.MustCompile(`(?s).*Missing value`),
				},
			},
		})
	})

	t.Run("error path - exist operator with value", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config: ResourceSubjectPatternRule("scc_spr", "Test Description",
						TestSubjectPatternRuleCondition{Variable: "name", Operator: "exist", Value: "unexpected"},
						defaultSubjectPattern,
					),
					ExpectError: regexp.MustCompile(`(?s).*Unexpected value`),
				},
			},
		})
	})

	t.Run("error path - exist_not operator with value", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config: ResourceSubjectPatternRule("scc_spr", "Test Description",
						TestSubjectPatternRuleCondition{Variable: "name", Operator: "exist_not", Value: "unexpected"},
						defaultSubjectPattern,
					),
					ExpectError: regexp.MustCompile(`(?s).*Unexpected value`),
				},
			},
		})
	})

	t.Run("error path - user_type is with invalid value", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config: ResourceSubjectPatternRule("scc_spr", "Test Description",
						TestSubjectPatternRuleCondition{Variable: "user_type", Operator: "is", Value: "Admin"},
						defaultSubjectPattern,
					),
					ExpectError: regexp.MustCompile(`(?s).*Invalid value`),
				},
			},
		})
	})

	t.Run("error path - user_type is_not with invalid value", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config: ResourceSubjectPatternRule("scc_spr", "Test Description",
						TestSubjectPatternRuleCondition{Variable: "user_type", Operator: "is_not", Value: "Guest"},
						defaultSubjectPattern,
					),
					ExpectError: regexp.MustCompile(`(?s).*Invalid value`),
				},
			},
		})
	})

	t.Run("error path - subject_pattern all fields empty", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config: ResourceSubjectPatternRule("scc_spr", "Test Description",
						defaultCondition,
						TestSubjectPattern{},
					),
					ExpectError: regexp.MustCompile(`(?s).*Empty subject pattern`),
				},
			},
		})
	})
}

func ResourceSubjectPatternRule(resourceName string, description string, condition TestSubjectPatternRuleCondition, subjectPattern TestSubjectPattern) string {
	return fmt.Sprintf(`
resource "scc_subject_pattern_rule" "%s" {
  description     = "%s"
  condition       = %s
  subject_pattern = %s
}`, resourceName, description, generateCondition(condition), generateSubjectPattern(subjectPattern))
}

func ResourceSubjectPatternRuleWoDescription(resourceName string, condition TestSubjectPatternRuleCondition, subjectPattern TestSubjectPattern) string {
	return fmt.Sprintf(`
resource "scc_subject_pattern_rule" "%s" {
  condition       = %s
  subject_pattern = %s
}`, resourceName, generateCondition(condition), generateSubjectPattern(subjectPattern))
}

func ResourceSubjectPatternRuleWoCondition(resourceName string, description string, subjectPattern TestSubjectPattern) string {
	return fmt.Sprintf(`
resource "scc_subject_pattern_rule" "%s" {
  description     = "%s"
  subject_pattern = %s
}`, resourceName, description, generateSubjectPattern(subjectPattern))
}

func ResourceSubjectPatternRuleWoSubjectPattern(resourceName string, description string, condition TestSubjectPatternRuleCondition) string {
	return fmt.Sprintf(`
resource "scc_subject_pattern_rule" "%s" {
  description = "%s"
  condition   = %s
}`, resourceName, description, generateCondition(condition))
}

// ResourceSubjectPatternRuleWoConditionVariable omits condition.variable to trigger a required-field error.
func ResourceSubjectPatternRuleWoConditionVariable(resourceName string, description string, condition TestSubjectPatternRuleCondition, subjectPattern TestSubjectPattern) string {
	return fmt.Sprintf(`
resource "scc_subject_pattern_rule" "%s" {
  description     = "%s"
  condition       = {
    operator = %q
    value    = %q
  }
  subject_pattern = %s
}`, resourceName, description, condition.Operator, condition.Value, generateSubjectPattern(subjectPattern))
}

// ResourceSubjectPatternRuleWoConditionOperator omits condition.operator to trigger a required-field error.
func ResourceSubjectPatternRuleWoConditionOperator(resourceName string, description string, condition TestSubjectPatternRuleCondition, subjectPattern TestSubjectPattern) string {
	return fmt.Sprintf(`
resource "scc_subject_pattern_rule" "%s" {
  description     = "%s"
  condition       = {
    variable = %q
    value    = %q
  }
  subject_pattern = %s
}`, resourceName, description, condition.Variable, condition.Value, generateSubjectPattern(subjectPattern))
}

func getImportStateForSubjectPatternRule(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return rs.Primary.Attributes["index"], nil
	}
}
