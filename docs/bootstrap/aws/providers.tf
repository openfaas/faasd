provider "aws" {
  version = "~> 3.6.0"
  region = var.aws_region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

provider "local" {
  version = "~> 1.4.0"
}

provider "random" {
  version = "~> 2.3.0"
}

provider "template" {
  version = "~> 2.1.2"
}

