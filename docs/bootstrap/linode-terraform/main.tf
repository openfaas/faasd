locals {
  generate_password   = var.basic_auth_password == null || var.basic_auth_password == ""
  basic_auth_user     = var.basic_auth_user
  basic_auth_password = local.generate_password ? random_password.faasd[0].result : var.basic_auth_password

  stackscript_data = {
    basic_auth_user     = local.basic_auth_user
    basic_auth_password = local.basic_auth_password
    domain              = var.domain
    email               = var.email
  }
}

resource "linode_sshkey" "ssh-key" {
  label   = "faasd-sshkey"
  ssh_key = chomp(file("~/.ssh/id_rsa.pub"))
}

resource "random_password" "faasd" {
  count            = local.generate_password ? 1 : 0
  length           = 16
  special          = true
  override_special = "_-#"
}


resource "linode_stackscript" "faasd_script" {
  description = "stackscript for faasd installtion"
  label       = "faasd-stackscript"
  images      = [var.instance_image]
  script      = templatefile("./templates/startup.sh", local.stackscript_data)
}

resource "linode_instance" "faasd_instance" {
  label           = var.instance_label
  image           = var.instance_image
  region          = var.region
  type            = var.instance_type
  authorized_keys = [linode_sshkey.ssh-key.ssh_key]
  root_pass       = var.root_pass
  stackscript_id  = linode_stackscript.faasd_script.id
}