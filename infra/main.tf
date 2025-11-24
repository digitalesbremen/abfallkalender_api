##
## Infrastructure main file (resources/data/locals only)
##
## This file intentionally contains only resources, data sources and locals.
## Version and provider constraints live in versions.tf, variables in variables.tf,
## and outputs in outputs.tf.

# ECR repository (private)
resource "aws_ecr_repository" "repo" {
  name                  = "abfallkalender-api"
  image_tag_mutability  = "MUTABLE"
  force_delete          = true # Caution: also deletes the repo when images exist (useful for dev/test)
  image_scanning_configuration {
    scan_on_push = true
  }
}

# ECR lifecycle policy – keep only the last 20 images (all tags)
resource "aws_ecr_lifecycle_policy" "policy" {
  repository = aws_ecr_repository.repo.name
  policy     = jsonencode({
    rules = [
      {
        rulePriority = 1,
        description  = "Keep last 20 images",
        selection    = {
          tagStatus   = "any",
          countType   = "imageCountMoreThan",
          countNumber = 20
        },
        action = { type = "expire" }
      }
    ]
  })
}

# GitHub OIDC provider (for GitHub Actions without long-lived keys)
# Note: If it already exists in the account, import it into state first:
#   tofu import aws_iam_openid_connect_provider.github arn:aws:iam::<ACCOUNT_ID>:oidc-provider/token.actions.githubusercontent.com
resource "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"
  client_id_list = [
    "sts.amazonaws.com"
  ]
  # GitHub OIDC root CA thumbprints as per GitHub documentation
  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1", # DigiCert Global Root CA
    "1c58a3a8518e8759bf075b76b750d4f2df264fcd"  # DigiCert Global Root G2
  ]
}

data "aws_caller_identity" "current" {}

# IAM role that GitHub Actions (only tag refs in this repo) may assume
resource "aws_iam_role" "github_actions_ecr_push" {
  name = "github-actions-ecr-push"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Principal = {
          Federated = aws_iam_openid_connect_provider.github.arn
        },
        Action = "sts:AssumeRoleWithWebIdentity",
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" : "sts.amazonaws.com"
          },
          StringLike = {
            # Allow only this repo and tag refs
            "token.actions.githubusercontent.com:sub" : "repo:digitalesbremen/abfallkalender_api:ref:refs/tags/*"
          }
        }
      }
    ]
  })
}

# Minimal permissions for ECR push
resource "aws_iam_policy" "ecr_push" {
  name        = "GitHubActionsECRPush"
  description = "Minimal permissions for pushing images to ECR"
  policy      = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:PutImage",
          "ecr:DescribeRepositories",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer"
        ],
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "attach_ecr_push" {
  role       = aws_iam_role.github_actions_ecr_push.name
  policy_arn = aws_iam_policy.ecr_push.arn
}

# ------------------------------------------------------------
# Lambda: execution role, log group, function, function URL
# ------------------------------------------------------------

# IAM role for Lambda execution (logs, etc.)
resource "aws_iam_role" "lambda_exec" {
  name = "abfallkalender-api-lambda-exec"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Principal = { Service = "lambda.amazonaws.com" },
        Action   = "sts:AssumeRole"
      }
    ]
  })
}

# Attach basic logging permissions
resource "aws_iam_role_policy_attachment" "lambda_basic_logs" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

locals {
  image_uri = "${aws_ecr_repository.repo.repository_url}:${var.image_tag}"
}

# Optional explicit log group (to control retention)
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/abfallkalender-api"
  retention_in_days = 14
}

# Lambda function from container image (AWS Lambda Web Adapter included in the image)
resource "aws_lambda_function" "app" {
  function_name = "abfallkalender-api"
  role          = aws_iam_role.lambda_exec.arn
  package_type  = "Image"
  image_uri     = local.image_uri

  architectures = ["arm64"]
  timeout       = var.lambda_timeout_s
  memory_size   = var.lambda_memory_mb

  # Optional hard cap for cost control – only set if explicitly desired.
  # When null, no reservation is set (recommended to avoid account quota issues).
  reserved_concurrent_executions = var.reserved_concurrency

  # The Go binary (HTTP server) is started via AWS Lambda Web Adapter.
  # No handler/runtime needed because package_type = Image.

  depends_on = [aws_cloudwatch_log_group.lambda]
}

# Public Function URL (no API Gateway)
resource "aws_lambda_function_url" "app_url" {
  function_name      = aws_lambda_function.app.function_name
  authorization_type = "NONE" # public; tighten later if needed

  cors {
    allow_credentials = false
    allow_headers     = ["*"]
    # AWS Lambda Function URL CORS does not allow "OPTIONS" in allow_methods.
    # It is handled automatically by the service. Otherwise you get:
    # "Value '[GET, HEAD, OPTIONS]' at 'cors.allowMethods' ... length <= 6" ("OPTIONS" = 7)
    allow_methods     = ["GET", "HEAD"]
    allow_origins     = ["*"]
    expose_headers    = []
    max_age           = 86400
  }
}

# --------------------
# Warmup via EventBridge
# --------------------
resource "aws_cloudwatch_event_rule" "warmup" {
  name                = "abfallkalender-api-warmup"
  description         = "Periodic warmup of the Lambda function"
  schedule_expression = "rate(5 minutes)"
}

resource "aws_cloudwatch_event_target" "warmup_target" {
  rule = aws_cloudwatch_event_rule.warmup.name
  arn  = aws_lambda_function.app.arn

  input = jsonencode({
    source = "warmup",
    action = "ping"
  })
}

# Permission so EventBridge can invoke the Lambda function
resource "aws_lambda_permission" "allow_events_warmup" {
  statement_id  = "AllowExecutionFromEventBridgeWarmup"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.app.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.warmup.arn
}
