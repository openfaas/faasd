output "instance_ip" {
  value = aws_eip.faasd.public_ip
}

output "gateway_url" {
  value = local.ip_url
}

output "password" {
  value = random_password.password.result
}

output "login_cmd" {
  value = "faas-cli login -g ${local.ip_url} -p ${random_password.password.result}"
}
