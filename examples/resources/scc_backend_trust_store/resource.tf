# Add a backend trust store certificate from a PEM file.
# The certificate must be in PEM format.
resource "scc_backend_trust_store" "scc_bts" {
  certificate = file("${path.module}/certs/certificate.pem")
}

# Add a backend trust store certificate by providing the PEM
# certificate content directly as a string.
resource "scc_backend_trust_store" "scc_bts" {
  certificate = <<-EOT
-----BEGIN CERTIFICATE-----
MIID...
...
-----END CERTIFICATE-----
EOT
}