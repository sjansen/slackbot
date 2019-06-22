resource "aws_cloudwatch_log_group" "fn" {
  name              = "/aws/lambda/${aws_lambda_function.fn.function_name}"
  retention_in_days = 14
}
