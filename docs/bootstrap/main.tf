terraform {
  required_version = ">= 1.0.4"
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "2.11.0"
    }
  }
}

variable "do_token" {}

variable "ssh_key_file" {
  default     = "~/.ssh/id_rsa.pub"
  description = "Path to the SSH public key file"
}

provider "digitalocean" {
  token = var.do_token
}

resource "random_password" "password" {
  length           = 16
  special          = true
  override_special = "_-#"
}

data "template_file" "cloud_init" {
  template = file("cloud-config.tpl")
  vars = {
    gw_password = random_password.password.result
  }
}

resource "digitalocean_ssh_key" "faasd_ssh_key" {
  name       = "ssh-key"
  public_key = file(var.ssh_key_file)
}

resource "digitalocean_droplet" "faasd" {

  region = "lon1"
  image  = "ubuntu-18-04-x64"
  name   = "faasd"
  # Plans: https://developers.digitalocean.com/documentation/changelog/api-v2/new-size-slugs-for-droplet-plan-changes/
  #size   = "512mb"
  size      = "s-1vcpu-1gb"
  user_data = data.template_file.cloud_init.rendered
  ssh_keys = [
    digitalocean_ssh_key.faasd_ssh_key.id
  ]
}

output "password" {
  value     = random_password.password.result
  sensitive = true
}

output "gateway_url" {
  value = "http://${digitalocean_droplet.faasd.ipv4_address}:8080/"
}

output "login_cmd" {
  value     = "faas-cli login -g http://${digitalocean_droplet.faasd.ipv4_address}:8080/ -p ${random_password.password.result}"
  sensitive = true
}

