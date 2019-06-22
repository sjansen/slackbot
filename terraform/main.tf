terraform {
  required_version = ">= 0.12.2"
}

##
# Credentials
##

provider "archive" {
  version = "~> 1.2.2"
}

provider "aws" {
  version = "~> 2.16.0"

  profile = "${var.aws_profile}"
  region = "${var.aws_region}"
}

provider "aws" {
  version = "~> 2.16"

  alias = "cloudfront"
  profile = "${var.aws_profile}"
  region = "us-east-1"
}

variable "aws_profile" {
  type = "string"
}

variable "aws_region" {
  default = "us-east-1"
}

##
# Resources
##

variable "dns_name" {
  type = "string"
}

variable "dns_zone" {
  type = "string"
}

variable "fn" {
  type = "string"
}

variable "publish_fn" {
  default = false
}

variable "lb" {
  type = "string"
}

variable "protect_lb" {
  default = false
}

variable "logs" {
  type = "string"
}

variable "media" {
  type = "string"
}

variable "protect_logs" {
  default = false
}

variable "protect_media" {
  default = false
}

variable "sg" {
  type = "string"
}

variable "slackbot_oauth_access_token" {
  type = "string"
}

variable "slackbot_verification_token" {
  type = "string"
}

variable "subnet_ids" {
  type = "list"
}

variable "use_alb" {
  default = true
}

variable "vpc_id" {
  type = "string"
}
