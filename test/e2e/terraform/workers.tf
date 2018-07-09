resource "oneandone_server" "worker" {
    count = "${var.num_workers}"
    name = "${data.null_data_source.cluster.outputs["name"]}-worker-${count.index + 1}"
    description = "Kubernetes worker"
    image = "CoreOS_Stable_64std"
    datacenter = "${var.region}"
    fixed_instance_size = "${data.oneandone_instance_size.M.id}"
    firewall_policy_id = "${oneandone_firewall_policy.fw.id}"
    ssh_key_public = "${tls_private_key.sshkey.public_key_openssh}"
}

output "workers" {
    value = "${
        map(
        "hostnames", "${oneandone_server.worker.*.name}",
        "ips", "${flatten(oneandone_server.worker.*.ips)}"
        )
    }"
}