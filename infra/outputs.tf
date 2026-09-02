output "aws_region" {
  description = "AWS region containing the workload resources."
  value       = var.aws_region
}

output "access_key_id" {
  description = "AWS access key ID for the sync workload."
  value       = module.sync_credentials.access_key_id
}

output "secret_access_key" {
  description = "AWS secret access key for the sync workload. Store in 1Password immediately."
  value       = module.sync_credentials.secret_access_key
  sensitive   = true
}

output "iam_user_arn" {
  description = "ARN of the dedicated IAM user."
  value       = module.sync_credentials.iam_user_arn
}

output "s3_bucket_name" {
  description = "Name of the S3 bucket created for the workload."
  value       = module.sync_credentials.s3_bucket_name
}

output "sqs_queue_url" {
  description = "URL of the SQS queue created for the workload."
  value       = module.sync_credentials.sqs_queue_url
}

output "sqs_queue_arn" {
  description = "ARN of the SQS queue created for the workload."
  value       = module.sync_credentials.sqs_queue_arn
}

output "sqs_dlq_url" {
  description = "URL of the SQS dead-letter queue."
  value       = module.sync_credentials.sqs_dlq_url
}

output "sqs_dlq_arn" {
  description = "ARN of the SQS dead-letter queue."
  value       = module.sync_credentials.sqs_dlq_arn
}
