# readest-to-obsidian-sync

Terraform for the AWS resources and credentials used by the Readest-to-Obsidian sync workload.

It creates a private S3 bucket, an encrypted SQS queue with notifications for newly created objects, one IAM user, an access key, and a least-privilege policy that permits:

- Receiving, deleting, and extending visibility of messages from one SQS queue.
- Listing a chosen prefix in one S3 bucket.
- Reading and writing objects under that S3 prefix.

## Usage

1. Configure an encrypted, access-controlled remote Terraform backend before applying. Terraform state contains the generated AWS secret access key.
2. Copy `infra/terraform.tfvars.example` to `infra/terraform.tfvars` and supply globally unique bucket and queue names.
3. From `infra`, run:

   ```sh
   make init
   make plan
   make apply
   ```

   The Makefile explicitly loads `terraform.tfvars`. Use `make plan TFVARS=another.tfvars` or the equivalent `make apply TFVARS=another.tfvars` to select a different variable file.

4. Store the `access_key_id` and `secret_access_key` outputs in 1Password. Create an ExternalSecret for the consuming Kubernetes workload from those values.

The generated key is not committed to this repository. Rotate it with `terraform apply -replace='module.sync_credentials.aws_iam_access_key.this'`, update 1Password, and let External Secrets reconcile before disabling the previous key.

## Architecture

Readest uploads state to S3. S3 sends object-created notifications to SQS. KEDA scales a Kubernetes `ScaledJob` from the queue, and the `readest-obsidian-sync run` worker receives, parses, logs, and acknowledges S3 events. This milestone does not reconcile Readest configuration or write to Obsidian.

This repository owns application code, container builds, Helm templates, and the Terraform definitions for the supporting AWS infrastructure. The actual Kubernetes deployment and environment-specific Helm values live in a separate Kubernetes/GitOps repository. Nothing in this repository deploys to a cluster.

## Worker

```sh
go test ./...
go vet ./...
go build ./...
go run ./cmd/readest-obsidian-sync run
```

The worker requires `AWS_REGION`, `SQS_QUEUE_URL`, and `S3_BUCKET`. It uses the AWS SDK default credential chain, long polls in batches of up to ten messages, logs structured JSON, deletes valid messages, and exits after an empty poll. It does not log credentials or object contents.

An S3 notification message has a `Records` array containing `eventSource: "aws:s3"`, `eventName`, `eventTime`, and nested bucket/object fields. The worker logs one event per record, including URL-decoded object keys and object size where available. Duplicate and out-of-order notifications are logged independently and are safe because this milestone performs no state mutation.

## Container

```sh
docker build --tag readest-obsidian-sync:0.1.0 .
```

The multi-stage image contains only the statically linked worker and runs as a non-root user. The intended entrypoint is `readest-obsidian-sync run`.

## Helm

The reusable chart at `charts/readest-obsidian-sync` renders a KEDA `ScaledJob` and ServiceAccount. It intentionally does not render a Service or Ingress. Defaults use one maximum replica because later reconciliation will target a single vault.

```sh
helm lint charts/readest-obsidian-sync
helm template test charts/readest-obsidian-sync \
  --set aws.region=eu-west-2 \
  --set aws.sqsQueueUrl=https://example.invalid/queue \
  --set aws.s3Bucket=example
```

The deployment repository supplies `aws.region`, `aws.sqsQueueUrl`, `aws.s3Bucket`, the image tag, and optionally `existingSecret`. An existing Secret must contain `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`. ServiceAccount annotations are available for future workload identity. No credentials or AWS account IDs are embedded in the chart.

## AWS Permissions

The Terraform module creates a standard encrypted SQS queue and permits the S3 service to send notifications only from the configured bucket and account. The worker IAM user can receive, delete, inspect, and extend visibility on that queue and can list/read/write/delete objects under the configured S3 prefix. Queue URL and bucket name are exposed as non-secret outputs; access key outputs remain sensitive.

## Validation And Releases

Run `make test`, `make vet`, `make build`, `make helm-lint`, and `make helm-template` locally. GitHub Actions repeats Go, Docker, Helm, and Terraform validation. Version `0.1.0` is used by the initial binary/image/chart contract; deployment consumers should pin image and chart versions rather than use `latest`.

The release workflow runs after pushes to `main`. It publishes a multi-architecture image to `ghcr.io/<owner>/readest-obsidian-sync:0.1.0-main.<commit-sha>` and the matching Helm chart to `oci://ghcr.io/<owner>/charts`. It authenticates with the repository-provided `GITHUB_TOKEN`; no publishing credentials are stored here.
