locals {
  ip_url = "http://${aws_eip.faasd.public_ip}:8080/"
  open_ports = [ 22, 80, 443, 8080 ]
}
