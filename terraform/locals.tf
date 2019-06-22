locals {
  apigw_url_parts = "${split("/",aws_api_gateway_deployment.default[0].invoke_url)}"
}
