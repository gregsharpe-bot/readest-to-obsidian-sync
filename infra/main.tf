terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

module "sync_credentials" {
  source = "./modules/s3-sqs-iam"

  user_name                = var.user_name
  sqs_queue_arn            = var.sqs_queue_arn
  s3_bucket_arn            = var.s3_bucket_arn
  s3_prefix                = var.s3_prefix
  permissions_boundary_arn = var.permissions_boundary_arn
  tags                     = var.tags
}
