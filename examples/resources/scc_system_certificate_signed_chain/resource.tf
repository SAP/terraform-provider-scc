resource "scc_system_certificate_signed_chain" "system_cert_signed_chain_as_file" {
  signed_chain = file("${path.module}/certs/signed_chain.pem")
}