variable "aws_access_key" {}

variable "aws_secret_key" {}

variable "aws_region" {
  description = "Region to deploy to"
  default = "eu-west-2" # London
}

variable "aws_ec2_instance_type" {
  description = "Instance type for the EC2 instance"
  default = "t2.micro" # Free tier
}

variable "ssh_key_file" {
  default     = "~/.ssh/id_rsa.pub"
  description = "Path to the SSH public key file"
}
