resource "oneandone_firewall_policy" "fw" {
  name = "${data.null_data_source.cluster.outputs["name"]}-fw"
  rules = [
    {
        "protocol" = "TCP"
        "port_from" = 22
        "port_to" = 22
        "source_ip" = "0.0.0.0"
    },
    {
        "protocol" = "TCP"
        "port_from" = 80
        "port_to" = 80
        "source_ip" = "0.0.0.0"
    },
    {
        "protocol" = "TCP"
        "port_from" = 443
        "port_to" = 443
        "source_ip" = "0.0.0.0"
    },
    {
        "protocol" = "TCP"
        "port_from" = 6443
        "port_to" = 6443
        "source_ip" = "0.0.0.0"
    },
    {
        "protocol" = "TCP"
        "port_from" = 12443
        "port_to" = 12443
        "source_ip" = "0.0.0.0"
    },
    {
        "protocol" = "TCP"
        "port_from" = 4001
        "port_to" = 4001
        "source_ip" = "192.168.100.0/24"
    },
    {
        "protocol" = "TCP"
        "port_from" = 30000
        "port_to" = 32767
        "source_ip" = "0.0.0.0"
    },
    {
        "protocol" = "TCP"
        "port_from" = 2379
        "port_to" = 2379
        "source_ip" = "192.168.100.0/24"
    },
    {
        "protocol" = "TCP"
        "port_from" = 2380
        "port_to" = 2380
        "source_ip" = "192.168.100.0/24"
    },
    {
        "protocol" = "TCP"
        "port_from" = 10250
        "port_to" = 10250
        "source_ip" = "192.168.100.0/24"
    },
    {
        "protocol" = "UDP"
        "port_from" = 2379
        "port_to" = 2379
        "source_ip" = "192.168.100.0/24"
    },
    {
        "protocol" = "UDP"
        "port_from" = 2380
        "port_to" = 2380
        "source_ip" = "192.168.100.0/24"
    },
    {
        "protocol" = "UDP"
        "port_from" = 8472
        "port_to" = 8472
        "source_ip" = "192.168.100.0/24"
    }
  ]
}
