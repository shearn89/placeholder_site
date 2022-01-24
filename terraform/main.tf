terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
}

provider "aws" {
  region = "eu-west-2"
}

resource "aws_acm_certificate" "cert" {
  domain_name = "jenkins.ajshearn.com"
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
    prevent_destroy = true
  }
}

resource "aws_s3_bucket" "log_bucket" {
  #checkov:skip=CKV_AWS_18 - logging
  #checkov:skip=CKV_AWS_21 - versioning
  #checkov:skip=CKV_AWS_144 - cross-region
  #checkov:skip=CKV_AWS_145 - kms
  bucket = "ajshearn-bucket-access-logs"
  acl = "log-delivery-write"
  lifecycle {
    prevent_destroy = true
  }
  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
}

resource "aws_s3_bucket" "placeholder_bucket" {
  #checkov:skip=CKV_AWS_21 - versioning
  #checkov:skip=CKV_AWS_144 - cross-region
  #checkov:skip=CKV_AWS_145 - kms
  bucket = "ajshearn.com"
  acl = "private"
  versioning {
    enabled = true
  }
  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
  logging {
    target_bucket = aws_s3_bucket.log_bucket.id
    target_prefix = "/s3/placeholder_bucket"
  }
  lifecycle_rule {
    id = "log_retention"
    enabled = true
    abort_incomplete_multipart_upload_days = 1
    expiration {
      days = 30
    }
    noncurrent_version_expiration {
      days = 7
    }
  }
}

resource "aws_cloudfront_origin_access_identity" "oai" {
  comment = "used for the placeholder site only"
}

resource "aws_cloudfront_distribution" "distribution" {
  #checkov:skip=CKV_AWS_68 - skip waf for cost
  origin {
    domain_name = aws_s3_bucket.placeholder_bucket.bucket_regional_domain_name
    origin_id = aws_s3_bucket.placeholder_bucket.bucket
    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.oai
    }
  }
  enabled = true
  comment = "CDN for the placeholder site"
  default_root_object = "index.html"
  aliases = ["ajshearn.com", "www.ajshearn.com"]
  logging_config {
    include_cookies = false
    bucket = aws_s3_bucket.log_bucket.id
    prefix = "/cloudfront/placeholder_site/"
  }
  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = aws_s3_bucket.placeholder_bucket.bucket
    viewer_protocol_policy = "redirect-to-https"
  }
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }
  viewer_certificate {
    acm_certificate_arn = aws_acm_certificate.cert.arn
    minimum_protocol_version = "TLSv1.2_2018"
  }
}

resource "aws_route53_zone" "hostedZone" {
  name = "ajshearn.com"
}

resource "aws_route53_record" "apex" {
  zone_id = aws_route53_zone.hostedZone.zone_id
  name    = "ajshearn.com"
  type    = "A"
  alias {
    name = aws_cloudfront_distribution.distribution.domain_name
    zone_id = aws_cloudfront_distribution.distribution.hosted_zone_id
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.hostedZone.zone_id
  name    = "www.ajshearn.com"
  type    = "A"
  alias {
    name = aws_cloudfront_distribution.distribution.domain_name
    zone_id = aws_cloudfront_distribution.distribution.hosted_zone_id
    evaluate_target_health = true
  }
}
