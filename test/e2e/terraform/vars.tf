variable "provider_token" {
  type = "string"
  description = "1&1 Cloud Panel API Key"
}

variable "cluster_name" {
  type = "string"
  description = "Cluster name, e.g. ccmtest1.  If not supplied, a name will be generated"
  default = ""
}

variable "region" {
  type = "string"
  default = "GB"
  description = "Datacentre"
}

variable "num_workers" {
  type = "string"
  default = "2"
  description = "The number of worker nodes in the cluster"
}
