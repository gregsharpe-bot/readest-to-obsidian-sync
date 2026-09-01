locals {
  normalized_prefix = trimsuffix(var.s3_prefix, "/")
  object_arn        = local.normalized_prefix == "" ? "${aws_s3_bucket.this.arn}/*" : "${aws_s3_bucket.this.arn}/${local.normalized_prefix}/*"
  list_prefixes     = local.normalized_prefix == "" ? ["*"] : ["${local.normalized_prefix}/*"]
}

resource "aws_s3_bucket" "this" {
  bucket = var.s3_bucket_name
  tags   = var.tags
}

resource "aws_s3_bucket_public_access_block" "this" {
  bucket = aws_s3_bucket.this.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "this" {
  bucket = aws_s3_bucket.this.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "this" {
  bucket = aws_s3_bucket.this.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_sqs_queue" "this" {
  name                    = var.sqs_queue_name
  sqs_managed_sse_enabled = true
  tags                    = var.tags
}

data "aws_iam_policy_document" "this" {
  statement {
    sid = "ConsumeQueueMessages"

    actions = [
      "sqs:ChangeMessageVisibility",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
      "sqs:ReceiveMessage",
    ]

    resources = [aws_sqs_queue.this.arn]
  }

  statement {
    sid       = "GetBucketLocation"
    actions   = ["s3:GetBucketLocation"]
    resources = [aws_s3_bucket.this.arn]
  }

  statement {
    sid = "ListPermittedObjects"

    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.this.arn]

    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = local.list_prefixes
    }
  }

  statement {
    sid = "ManagePermittedObjects"

    actions = [
      "s3:DeleteObject",
      "s3:GetObject",
      "s3:PutObject",
    ]

    resources = [local.object_arn]
  }
}

resource "aws_iam_user" "this" {
  name                 = var.user_name
  permissions_boundary = var.permissions_boundary_arn
  tags                 = var.tags
}

resource "aws_iam_user_policy" "this" {
  name   = "s3-sqs-sync"
  user   = aws_iam_user.this.name
  policy = data.aws_iam_policy_document.this.json
}

resource "aws_iam_access_key" "this" {
  user = aws_iam_user.this.name

  lifecycle {
    create_before_destroy = true
  }
}
