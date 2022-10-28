
terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

// note: terraform apply -var digitalocean_token=$DIGITALOCEAN_TOKEN

provider "digitalocean" {
  token = var.digitalocean_token
}

variable "digitalocean_token" {
  type      = string
  sensitive = true
}

variable "digitalocean_ssh_key" {
  type      = string
  sensitive = false
  default   = "phi-failure-demo"
}

data "digitalocean_ssh_key" "phi" {
  name = var.digitalocean_ssh_key
}

// phi-accrual server 
resource "digitalocean_droplet" "server" {
  name          = "server-node-nyc"
  image         = "ubuntu-22-04-x64"
  size          = "s-1vcpu-2gb"
  region        = "nyc3"
  droplet_agent = true
  monitoring    = true
  user_data     = file("${path.module}/../provisioning/server/provision.sh")
  ssh_keys      = [data.digitalocean_ssh_key.phi.id]
}

// phi-accrual test nodes
resource "digitalocean_droplet" "nodes" {
  for_each = toset(["nyc1", "fra1", "sgp1"])

  name          = "client-node-${each.key}"
  image         = "ubuntu-22-04-x64"
  size          = "s-1vcpu-2gb"
  region        = each.key
  droplet_agent = true
  monitoring    = true
  user_data = templatefile(
    "${path.module}/../provisioning/client/provision.sh", {
      FAILURE_DETECTOR_SERVER_HOST = digitalocean_droplet.server.ipv4_address
    }
  )
  ssh_keys = [data.digitalocean_ssh_key.phi.id]
}


output "server_hostname" {
  value = digitalocean_droplet.server.ipv4_address
}

output "node_hostnames" {
  value = [for n in digitalocean_droplet.nodes : n.ipv4_address]
}