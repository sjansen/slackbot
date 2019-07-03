data "archive_file" "edge" {
  type        = "zip"
  output_path = "../dist/edge.zip"
  source {
    filename = "index.js"
    content  = file("../edge/rewrite.js")
  }
}


resource "aws_lambda_function" "edge" {
  provider = "aws.cloudfront"

  function_name    = "${var.fn}-rewrite"
  filename         = data.archive_file.edge.output_path
  source_code_hash = data.archive_file.edge.output_base64sha256
  role             = aws_iam_role.edge.arn
  runtime          = "nodejs8.10"
  handler          = "index.handler"
  memory_size      = 128
  timeout          = 3
  publish          = true
}


resource "aws_lambda_function" "fn" {
  function_name    = var.fn
  filename         = "../dist/slackbot.zip"
  handler          = "slackbot"
  source_code_hash = filebase64sha256("../dist/slackbot.zip")
  role             = aws_iam_role.fn.arn

  runtime     = "go1.x"
  memory_size = 128
  timeout     = 15

  environment {
    variables = {
      SLACKBOT_USE_ALB            = var.use_alb ? "true" : "false"
      SLACKBOT_DYNAMODB_TABLE     = var.db
      SLACKBOT_OAUTH_ACCESS_TOKEN = var.slackbot_oauth_access_token
      SLACKBOT_VERIFICATION_TOKEN = var.slackbot_verification_token
    }
  }

  tracing_config {
    mode = "Active"
  }
}


resource "aws_lambda_permission" "lb" {
  count = var.use_alb ? 1 : 0

  statement_id  = "AllowExecutionFromALB"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.fn.arn
  principal     = "elasticloadbalancing.amazonaws.com"
  source_arn    = join("", aws_alb_target_group.slackbot.*.arn)
}


resource "aws_lambda_permission" "apigw" {
  count = var.use_alb ? 0 : 1

  statement_id  = "AllowExecutionFromAPIGW"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.fn.arn
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${join("", aws_api_gateway_deployment.default.*.execution_arn)}*/*/*"
}
