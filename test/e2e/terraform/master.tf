resource "oneandone_server" "master" {
  name = "${data.null_data_source.cluster.outputs["name"]}-master-1"
  description = "Kubernetes master"
  image = "CoreOS_Stable_64std"
  datacenter = "${var.region}"
  fixed_instance_size = "${data.oneandone_instance_size.L.id}"
  firewall_policy_id = "${oneandone_firewall_policy.fw.id}"
  ssh_key_public = "${tls_private_key.sshkey.public_key_openssh}"
}

output "master" {
  value = "${
    map(
      "hostname", "${oneandone_server.master.name}",
      "ip", "${oneandone_server.master.ips.0.ip}"
    )
  }"
}