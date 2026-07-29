# This feature requires Terraform v1.14.0 or later.
# List resource queries must be defined in .tfquery.hcl files and run with:
#   terraform query

# ---------------------------------------------------------------------------
# Example 1: Fetch all subject pattern rules
# Omit the index filter to return every rule configured in the SCC instance.
# ---------------------------------------------------------------------------
list "scc_subject_pattern_rule" "all" {
  provider = scc
}

# ---------------------------------------------------------------------------
# Example 2: Fetch a single rule by its positional index
# Rules are zero-based. Index 0 is the first rule.
# ---------------------------------------------------------------------------
list "scc_subject_pattern_rule" "by_index" {
  provider = scc

  config {
    index = 0
  }
}

# ---------------------------------------------------------------------------
# Example 3: Fetch a single rule with full resource data
# Set include_resource = true to also return description, condition, and
# subject_pattern fields alongside the identity.
# ---------------------------------------------------------------------------
list "scc_subject_pattern_rule" "by_index_with_data" {
  provider         = scc
  include_resource = true

  config {
    index = 0
  }
}
