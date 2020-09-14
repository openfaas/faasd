data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "random_password" "password" {
  length = 16
  special = true
  override_special = "_-#"
}

data "local_file" "ssh_key"{
  filename = pathexpand(var.ssh_key_file)
}

data "template_file" "cloud_init" {
  template = file("cloud-config.sh")
  vars = {
    gw_password = random_password.password.result
    ssh_key = data.local_file.ssh_key.content
//    enable_custom_domain = var.custom_domain_enabled
//    faasd_domain_name = var.custom_domain_enabled ? local.custom_domain : null # Creates circular dependency if IP injected
//    letsencrypt_email = var.letsencrypt_email
  }
}

resource "aws_key_pair" "faasd" {
  key_name = "faasd"
  public_key = data.local_file.ssh_key.content
}

resource "aws_vpc" "faasd" {
  cidr_block = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support = true
}

resource "aws_subnet" "faasd" {
  cidr_block = cidrsubnet(aws_vpc.faasd.cidr_block, 3, 1)
  vpc_id = aws_vpc.faasd.id
  availability_zone = "${var.aws_region}a" # May be a bit flakey - good enough for bootstrapping
}

resource "aws_security_group" "faasd" {
  name = "allow-all-faasd"
  description = "Allow all incoming traffic"
  vpc_id = aws_vpc.faasd.id

  dynamic "ingress" {
    for_each = local.open_ports
    content {
      cidr_blocks = [
        "0.0.0.0/0"
      ]
      from_port = ingress.value
      to_port = ingress.value
      protocol = "tcp"
    }
  }

  // Terraform removes the default rule
  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "faasd" {
  ami = data.aws_ami.ubuntu.id
  instance_type = var.aws_ec2_instance_type
  user_data_base64 = base64encode(data.template_file.cloud_init.rendered)
  key_name = aws_key_pair.faasd.key_name
  security_groups = [aws_security_group.faasd.id]
  subnet_id = aws_subnet.faasd.id

  tags = {
    Name = "faasd"
  }
}

resource "aws_eip" "faasd" {
  instance = aws_instance.faasd.id
  vpc = true
}

resource "aws_internet_gateway" "faasd" {
  vpc_id = aws_vpc.faasd.id

  tags = {
    Name = "faasd"
  }
}

resource "aws_route_table" "faasd" {
  vpc_id = aws_vpc.faasd.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.faasd.id
  }

  tags = {
    Name = "faasd"
  }
}

resource "aws_route_table_association" "faasd" {
  route_table_id = aws_route_table.faasd.id
  subnet_id = aws_subnet.faasd.id
}
