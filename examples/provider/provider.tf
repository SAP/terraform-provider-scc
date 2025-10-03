terraform {
  required_providers {
    scc = {
        source = "sap/scc"
        version = "1.0.0-rc1"
    }
  }
}

provider "scc" {
  # Cloud Connector base URL, e.g., "https://localhost:8443"
  instance_url = "https://your-cloud-connector-instance:8443"

  # ❗ Authentication: Use **either** Basic Auth or Client Certificate Auth — not both

  # Option 1: Basic Authentication (set username and password)
  username = var.scc_username              # or set via SCC_USERNAME
  password = var.scc_password              # or set via SCC_PASSWORD (Sensitive)

  # Option 2: Certificate-based Authentication (set both client_certificate and client_key)
  # client_certificate = file("${path.module}/certs/client.crt")   # or SCC_CLIENT_CERTIFICATE
  # client_key         = file("${path.module}/certs/client.key")   # or SCC_CLIENT_KEY

  # TLS Server Verification
  ca_certificate = file("${path.module}/certs/ca.pem")             # or SCC_CA_CERTIFICATE
}

