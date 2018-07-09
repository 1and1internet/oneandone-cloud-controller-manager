resource "oneandone_private_network" "pn" {
    name = "${data.null_data_source.cluster.outputs["name"]}-pn"
    description = "Private network for Kube cluster"
    datacenter = "GB"
    network_address = "192.168.100.0"
    subnet_mask = "255.255.255.0"
    server_ids = [
        "${oneandone_server.master.id}",
        "${oneandone_server.worker.*.id}"
    ]
}
