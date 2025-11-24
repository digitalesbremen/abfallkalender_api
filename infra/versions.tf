##
## versions.tf — Terraform/OpenTofu and provider constraints
##
## Managed with OpenTofu. The terraform block is understood by both Terraform
## and OpenTofu. Provider versions are pinned for repeatable builds.

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

provider "aws" {
  region = "eu-central-1"
}
