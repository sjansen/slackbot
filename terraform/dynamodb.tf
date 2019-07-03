resource "aws_dynamodb_table" "db" {
  name           = var.db
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  server_side_encryption {
    enabled = true
  }

  ttl {
    attribute_name = "expires"
    enabled        = true
  }
}
