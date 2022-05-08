output "ip_address" {
  description = "The public IP address of the faasd instance"
  value       = linode_instance.faasd_instance.ip_address
}

output "gateway_url" {
  description = "The url of the faasd gateway"
  value       = var.domain == null || var.domain == "" ? format("http://%s:8080", linode_instance.faasd_instance.ip_address) : format("https://%s", var.domain)
}

output "basic_auth_user" {
  description = "The basic auth user name."
  value       = var.basic_auth_user
}

output "basic_auth_password" {
  description = "The basic auth password."
  value       = "/var/lib/faasd/secrets/basic-auth-password"
}