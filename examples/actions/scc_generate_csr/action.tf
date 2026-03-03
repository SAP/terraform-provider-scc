action "scc_generate_csr" "system_cert" {
  config {
    type     = "system" # Options: ca | system | ui
    key_size = 2048     # Options: 2048 | 4096
    subject_dn = {
      cn    = "cloud-connector.example.com"
      o     = "SAP"
      ou    = "BTP Engineering and Delivery Services India"
      l     = "Bangalore"
      st    = "Karnataka"
      c     = "IN"
      email = "admin@example.com"
    }
    subject_alternative_names = [
      {
        type  = "DNS"
        value = "cloud-connector.example.com"
      },
      {
        type  = "DNS"
        value = "cc.internal.local"
      }
    ]
  }
}