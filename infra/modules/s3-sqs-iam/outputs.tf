output "access_key_id" {
  description = "AWS access key ID for the IAM user."
  value       = aws_iam_access_key.this.id
}

output "secret_access_key" {
  description = "AWS secret access key for the IAM user."
  value       = aws_iam_access_key.this.secret
  sensitive   = true
}

output "iam_user_arn" {
  description = "ARN of the dedicated IAM user."
  value       = aws_iam_user.this.arn
}
