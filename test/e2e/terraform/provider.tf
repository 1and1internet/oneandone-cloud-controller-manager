provider "oneandone"{
  version = "~> 1.1"
  token = "${var.provider_token}"
  retries = 50
}