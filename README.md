# readest-to-obsidian-sync

Terraform for the AWS credentials used by the Readest-to-Obsidian sync workload.

It creates one IAM user, an access key, and a least-privilege policy that permits:

- Receiving, deleting, and extending visibility of messages from one SQS queue.
- Listing a chosen prefix in one S3 bucket.
- Reading and writing objects under that S3 prefix.

## Usage

1. Configure an encrypted, access-controlled remote Terraform backend before applying. Terraform state contains the generated AWS secret access key.
2. Copy `infra/terraform.tfvars.example` to `infra/terraform.tfvars` and fill in the resource ARNs and names.
3. From `infra`, run:

   ```sh
   terraform init
   terraform plan
   terraform apply
   ```

4. Store the `access_key_id` and `secret_access_key` outputs in 1Password. Create an ExternalSecret for the consuming Kubernetes workload from those values.

The generated key is not committed to this repository. Rotate it with `terraform apply -replace='module.sync_credentials.aws_iam_access_key.this'`, update 1Password, and let External Secrets reconcile before disabling the previous key.
