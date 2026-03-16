resource "scc_ui_certificate_signed_chain" "ui_cert_signed_chain_as_file" {
  signed_chain = file("${path.module}/certs/signed_chain.pem")
}