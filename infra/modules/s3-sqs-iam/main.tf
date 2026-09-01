locals {
  normalized_prefix = trimsuffix(var.s3_prefix, "/")
  object_arn        = local.normalized_prefix == "" ? "${var.s3_bucket_arn}/*" : "${var.s3_bucket_arn}/${local.normalized_prefix}/*"
  list_prefixes     = local.normalized_prefix == "" ? ["*"] : ["${local.normalized_prefix}/*"]
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

    resources = [var.sqs_queue_arn]
  }

  statement {
    sid = "ListPermittedObjects"

    actions   = ["s3:ListBucket"]
    resources = [var.s3_bucket_arn]

    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = local.list_prefixes
    }
  }

  statement {
    sid = "ReadAndWritePermittedObjects"

    actions = [
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
