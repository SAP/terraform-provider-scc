# Allow only Business users whose login name is "jdoe"
resource "scc_subject_pattern_rule" "business_user" {
  description = "Allow business user jdoe"

  condition = {
    variable = "login_name"
    operator = "is"
    value    = "jdoe"
  }

  subject_pattern = {
    cn = "John Doe"
    o  = "ACME Corp"
    c  = "DE"
  }
}

# Block all Technical users
resource "scc_subject_pattern_rule" "block_technical" {
  description = "Block all technical users"

  condition = {
    variable = "user_type"
    operator = "is_not"
    value    = "Technical"
  }

  subject_pattern = {
    cn = "technical-user"
  }
}

# Allow any user where the email attribute exists
resource "scc_subject_pattern_rule" "email_exists" {
  description = "Allow users with email attribute"

  condition = {
    variable = "email"
    operator = "exist"
  }

  subject_pattern = {
    cn    = "Email User"
    email = "user@example.com"
    o     = "Example Corp"
    c     = "US"
  }
}
