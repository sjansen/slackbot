output "media_bucket" {
  value = aws_s3_bucket.media.id
}

output "url" {
  value = "https://${var.dns_name}/"
}
