resource "aws_ssm_parameter" "slackbot_oauth_access_token" {
  count = var.use_ssm ? 1 : 0

  name        = "/${var.slackbot_oauth_access_token}"
  description = "An OAuth access token used to make Slack API calls"
  type        = "SecureString"
  value       = "invalid"
  overwrite   = false

  lifecycle {
    ignore_changes = [value]
  }
}


resource "aws_ssm_parameter" "slackbot_req_signing_secret" {
  count = var.use_ssm ? 1 : 0

  name        = "/${var.slackbot_req_signing_secret}"
  description = "A secret to verify requests from Slack"
  type        = "SecureString"
  value       = "invalid"
  overwrite   = false

  lifecycle {
    ignore_changes = [value]
  }
}
