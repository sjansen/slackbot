data "aws_iam_policy_document" "edge" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type = "Service"
      identifiers = [
        "lambda.amazonaws.com",
        "edgelambda.amazonaws.com"
      ]
    }
  }
}


data "aws_iam_policy_document" "media" {
  statement {
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.media.arn}/*"]
    principals {
      type        = "AWS"
      identifiers = [aws_cloudfront_origin_access_identity.cdn.iam_arn]
    }
  }
}


resource "aws_iam_policy" "fn-db" {
  name   = "${var.fn}-db"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
       "Effect": "Allow",
       "Action": [
         "dynamodb:BatchGetItem",
         "dynamodb:BatchWriteItem",
         "dynamodb:GetItem",
         "dynamodb:PutItem",
         "dynamodb:Query",
         "dynamodb:UpdateItem"
       ],
       "Resource": "${aws_dynamodb_table.db.arn}"
    }
  ]
}
EOF
}


resource "aws_iam_policy" "fn-logs" {
  name = "${var.fn}-logs"
  path = "/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    }
  ]
}
EOF
}


resource "aws_iam_policy" "fn-ssm" {
  name = "${var.fn}-ssm"
  path = "/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:GetParameters"
      ],
      "Resource": [
        "arn:aws:ssm:*:*:parameter/${var.slackbot_oauth_access_token}",
        "arn:aws:ssm:*:*:parameter/${var.slackbot_verification_token}"
      ]
    }
  ]
}
EOF
}


resource "aws_iam_policy" "fn-xray" {
    name = "${var.fn}-xray"
    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": {
        "Effect": "Allow",
        "Action": [
            "xray:PutTraceSegments",
            "xray:PutTelemetryRecords"
        ],
        "Resource": [
            "*"
        ]
    }
}
EOF
}


resource "aws_iam_role" "edge" {
  name_prefix        = "${var.fn}-edge"
  assume_role_policy = data.aws_iam_policy_document.edge.json
}


resource "aws_iam_role" "fn" {
  name = var.fn

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      }
    }
  ]
}
EOF
}


resource "aws_iam_role_policy_attachment" "edge" {
  role = aws_iam_role.edge.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}


resource "aws_iam_role_policy_attachment" "fn-db" {
  policy_arn = aws_iam_policy.fn-db.arn
  role = aws_iam_role.fn.name
}


resource "aws_iam_role_policy_attachment" "fn-logs" {
  policy_arn = aws_iam_policy.fn-logs.arn
  role = aws_iam_role.fn.name
}


resource "aws_iam_role_policy_attachment" "fn-ssm" {
  policy_arn = aws_iam_policy.fn-ssm.arn
  role = aws_iam_role.fn.name
}


resource "aws_iam_role_policy_attachment" "fn-xray" {
  policy_arn = aws_iam_policy.fn-xray.arn
  role = aws_iam_role.fn.name
}
