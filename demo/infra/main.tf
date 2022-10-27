
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

// data: https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/data-sources/ssh_key
data "digitalocean_ssh_key" "phi" {
  name = var.digitalocean_ssh_key
}

// phi-accrual server 
// resource: https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/droplet
resource "digitalocean_droplet" "server" {
  name          = "server-node-nyc"
  image         = "docker-20-04"
  size          = "s-1vcpu-1gb"
  region        = "nyc3"
  droplet_agent = true
  monitoring    = true
  ssh_keys      = [data.digitalocean_ssh_key.phi.id]
}

// phi-accrual test nodes
// resource: https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/droplet
resource "digitalocean_droplet" "nodes" {
  for_each = toset(["nyc1", "fra1", "sgp1"])

  name          = "client-node-${each.key}"
  image         = "docker-20-04"
  size          = "s-1vcpu-1gb"
  region        = each.key
  droplet_agent = true
  monitoring    = true
  ssh_keys      = [data.digitalocean_ssh_key.phi.id]
}

output "digitalocean_ssh_keyname" {
  value = data.digitalocean_ssh_key.phi.name
}

output "server_hostname" {
  value = digitalocean_droplet.server.ipv4_address
}

output "node_hostnames" {
  value = [for n in digitalocean_droplet.nodes : n.ipv4_address]
}