variable "user_name" {
  description = "Name for the dedicated IAM user."
  type        = string
}

variable "sqs_queue_name" {
  description = "Name for the SQS queue created for the workload."
  type        = string
}

variable "s3_bucket_name" {
  description = "Globally unique name for the S3 bucket created for the workload."
  type        = string
}

variable "s3_prefix" {
  description = "Object prefix the workload can access. Empty permits the entire bucket."
  type        = string
  default     = ""
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
  default     = {}
}
