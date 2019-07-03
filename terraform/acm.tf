resource "aws_acm_certificate" "cert" {
  provider = "aws.cloudfront"

  domain_name       = var.dns_name
  validation_method = "DNS"
}


resource "aws_acm_certificate_validation" "cert" {
  provider = "aws.cloudfront"

  certificate_arn         = aws_acm_certificate.cert.arn
  validation_record_fqdns = [aws_route53_record.cert.fqdn]
}
