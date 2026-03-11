resource "scc_ca_certificate_signed_chain" "ca_cert_signed_chain_as_file" {
  signed_chain = file("${path.module}/certs/signed_chain.pem")
}