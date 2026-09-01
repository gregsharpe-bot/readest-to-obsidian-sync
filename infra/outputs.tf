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
