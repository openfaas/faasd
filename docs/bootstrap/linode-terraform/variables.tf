variable "api_token" {
  description = " Linode API token"
}

variable "basic_auth_password" {
  description = "The basic auth password, if left empty, a random password is generated."
  type        = string
  default     = null
  sensitive   = true
}

variable "domain" {
  description = "Your public domain"
  type        = string
}

variable "email" {
  description = "Email used to order a certificate from Letsencrypt"
  type        = string
}

variable "basic_auth_user" {
  description = "The basic auth user name."
  type        = string
}

variable "instance_image" {
  description = "Image to use for Linode instance."
  type        = string
}

variable "instance_label" {
  description = "The Linode's label is for display purposes only, but must be unique."
  type        = string
}

variable "region" {
  description = "The region where your Linode will be located."
  type        = string
}

variable "instance_type" {
  description = "Your Linode's plan type."
  type        = string
}

variable "root_pass" {
  description = "Your Linode's root user's password."
  type        = string
  sensitive   = true
}