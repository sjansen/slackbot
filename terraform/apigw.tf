resource "aws_api_gateway_rest_api" "gw" {
  count = var.use_alb ? 0 : 1

  name = var.lb
  endpoint_configuration {
    types = ["REGIONAL"]
  }
}


resource "aws_api_gateway_deployment" "default" {
  count = var.use_alb ? 0 : 1
  depends_on = [
    "aws_api_gateway_integration.public",
  ]

  rest_api_id = join("", aws_api_gateway_rest_api.gw.*.id)
  stage_name  = "default"
}


resource "aws_api_gateway_method" "public" {
  count = var.use_alb ? 0 : 1

  authorization = "NONE"
  rest_api_id   = join("", aws_api_gateway_rest_api.gw.*.id)
  resource_id   = join("", aws_api_gateway_resource.proxy.*.id)
  http_method   = "ANY"
}


resource "aws_api_gateway_integration" "public" {
  count = var.use_alb ? 0 : 1

  rest_api_id = join("", aws_api_gateway_rest_api.gw.*.id)
  resource_id = join("", aws_api_gateway_method.public.*.resource_id)
  http_method = join("", aws_api_gateway_method.public.*.http_method)

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.fn.invoke_arn
}


resource "aws_api_gateway_resource" "proxy" {
  count = var.use_alb ? 0 : 1

  rest_api_id = join("", aws_api_gateway_rest_api.gw.*.id)
  parent_id   = join("", aws_api_gateway_rest_api.gw.*.root_resource_id)
  path_part   = "{proxy+}"
}
