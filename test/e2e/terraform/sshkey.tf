resource "tls_private_key" "sshkey" {
  algorithm   = "RSA"
}

output "ssh_private_key" {
  value = "${tls_private_key.sshkey.private_key_pem}"
}