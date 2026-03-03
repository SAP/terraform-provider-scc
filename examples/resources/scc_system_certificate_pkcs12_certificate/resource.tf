resource "scc_system_certificate_pkcs12_certificate" "system_cert_p12_as_file" {
  pkcs12_certificate = filebase64("${path.module}/certs/certificate.p12")
  password           = "test"
  key_password       = "test"
}