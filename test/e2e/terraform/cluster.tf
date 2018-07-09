data "oneandone_instance_size" "M" {
  name = "M"
}

data "oneandone_instance_size" "L" {
  name = "L"
}

resource "random_pet" "cluster_name" {
}

data "null_data_source" "cluster" {
  inputs = {
    name = "${var.cluster_name == "" ? random_pet.cluster_name.id : var.cluster_name}"
  }
}

output "cluster_name" {
  value = "${data.null_data_source.cluster.outputs["name"]}"
}