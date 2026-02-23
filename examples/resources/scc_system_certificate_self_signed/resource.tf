resource "scc_system_certificate_self_signed" "self_signed_cert" {
  key_size = 2048
  subject_dn = {
    cn = "example.com"
  }
}