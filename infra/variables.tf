variable "aws_region" {
  description = "AWS region containing the SQS queue and S3 bucket."
  type        = string
}

variable "user_name" {
  description = "Name for the dedicated IAM user."
  type        = string
  default     = "readest-to-obsidian-sync"
}

variable "sqs_queue_arn" {
  description = "ARN of the SQS queue the workload consumes."
  type        = string
}

variable "s3_bucket_arn" {
  description = "ARN of the S3 bucket, without an object prefix."
  type        = string

  validation {
    condition     = can(regex("^arn:[^:]+:s3:::[^/]+$", var.s3_bucket_arn))
    error_message = "s3_bucket_arn must be a bucket ARN such as arn:aws:s3:::my-bucket."
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
