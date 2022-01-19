# Placeholder Site

This is a simple static website that can be used as a placeholder.

The techy bit is that it's all deployed via the CDK onto AWS - content in `web/` should be uploaded to your S3 bucket, then the CDK run with `cdk deploy` to wrap it in a CloudFront Distribution.

You'll also need a certificate issued via Certificate Manager.

Dig into the template for details.
