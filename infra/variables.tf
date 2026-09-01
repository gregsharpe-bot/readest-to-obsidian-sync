variable "aws_region" {
  description = "AWS region containing the SQS queue and S3 bucket."
  type        = string
}

variable "user_name" {
  description = "Name for the dedicated IAM user."
  type        = string
  default     = "readest-to-obsidian-sync"
}

variable "sqs_queue_name" {
  description = "Name for the SQS queue created for the workload."
  type        = string
}

variable "s3_bucket_name" {
  description = "Globally unique name for the S3 bucket created for the workload."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$", var.s3_bucket_name))
    error_message = "s3_bucket_name must be a valid 3-63 character S3 bucket name."
  }
}

variable "s3_prefix" {
  description = "Object prefix the workload can access. Use an empty string for the entire bucket."
  type        = string
  default     = ""

  validation {
    condition     = !startswith(var.s3_prefix, "/")
    error_message = "s3_prefix must not begin with a slash."
  }
}

variable "permissions_boundary_arn" {
  description = "Optional IAM permissions boundary ARN to attach to the user."
  type        = string
  default     = null
  nullable    = true
}

variable "tags" {
  description = "Tags to apply to IAM resources."
  type        = map(string)
  default = {
    ManagedBy = "Terraform"
  }
}
