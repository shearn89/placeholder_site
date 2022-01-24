# Placeholder Site

This is a simple static website that can be used as a placeholder.

The techy bit is that it's all deployed via Terraform onto AWS - content in `web/` should be uploaded to your S3 bucket, then run terraform with `terraform apply` to wrap it in a CloudFront Distribution.

Dig into the `terraform/main.tf` for details.
