terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

provider "aws" {
  region = "eu-central-1"
}

# -------------
# Variablen
# -------------
variable "image_tag" {
  type        = string
  description = "Tag des ECR-Images für die Lambda-Funktion (muss bereits im ECR existieren)."
}

variable "lambda_memory_mb" {
  type        = number
  default     = 512
  description = "Arbeitsspeicher der Lambda-Funktion in MB."
}

variable "lambda_timeout_s" {
  type        = number
  default     = 15
  description = "Timeout der Lambda-Funktion in Sekunden."
}

# ECR Repository (privat)
resource "aws_ecr_repository" "repo" {
  name                  = "abfallkalender-api"
  image_tag_mutability  = "MUTABLE"
  force_delete          = true # Achtung: löscht Repo auch mit Images (praktisch für Dev/Test)
  image_scanning_configuration {
    scan_on_push = true
  }
}

# ECR Lifecycle Policy – halte nur die letzten 20 Images (alle Tags)
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

# GitHub OIDC Provider (für GitHub Actions ohne Langzeit-Keys)
# Hinweis: Falls bereits im Account vorhanden, ggf. importieren:
#   tofu import aws_iam_openid_connect_provider.github arn:aws:iam::<ACCOUNT_ID>:oidc-provider/token.actions.githubusercontent.com
resource "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"
  client_id_list = [
    "sts.amazonaws.com"
  ]
  # GitHub OIDC Root CA Thumbprint laut GitHub-Dokumentation
  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1", # DigiCert Global Root CA
    "1c58a3a8518e8759bf075b76b750d4f2df264fcd"  # DigiCert Global Root G2
  ]
}

data "aws_caller_identity" "current" {}

# IAM Rolle, die GitHub Actions (nur für Tags in diesem Repo) annehmen darf
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
            # Erlaube nur das aktuelle Repo und Tag-Refs
            "token.actions.githubusercontent.com:sub" : "repo:digitalesbremen/abfallkalender_api:ref:refs/tags/*"
          }
        }
      }
    ]
  })
}

# Minimale Berechtigungen für ECR Push
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
# Lambda: IAM Ausführungsrolle, LogGroup, Funktion, Function URL
# ------------------------------------------------------------

# IAM Rolle für Lambda-Ausführung (Logs etc.)
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

# Standard Logging Rechte
resource "aws_iam_role_policy_attachment" "lambda_basic_logs" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

locals {
  image_uri = "${aws_ecr_repository.repo.repository_url}:${var.image_tag}"
}

# Optionale explizite Log Group (ermöglicht Retention-Steuerung)
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/abfallkalender-api"
  retention_in_days = 14
}

# Lambda-Funktion aus Container-Image (mit AWS Lambda Web Adapter im Image enthalten)
resource "aws_lambda_function" "app" {
  function_name = "abfallkalender-api"
  role          = aws_iam_role.lambda_exec.arn
  package_type  = "Image"
  image_uri     = local.image_uri

  architectures = ["arm64"]
  timeout       = var.lambda_timeout_s
  memory_size   = var.lambda_memory_mb

  # Das Go-Binary (HTTP-Server) wird via AWS Lambda Web Adapter gestartet.
  # Kein handler/runtime nötig, da package_type = Image.

  depends_on = [aws_cloudwatch_log_group.lambda]
}

# Öffentliche URL direkt an der Funktion (ohne API Gateway)
resource "aws_lambda_function_url" "app_url" {
  function_name      = aws_lambda_function.app.function_name
  authorization_type = "NONE" # öffentlich; ggf. später absichern

  cors {
    allow_credentials = false
    allow_headers     = ["*"]
    allow_methods     = ["GET", "HEAD", "OPTIONS"]
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
  description         = "Periodisches Warmup der Lambda-Funktion"
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

# Berechtigung, damit EventBridge die Lambda-Funktion aufrufen darf
resource "aws_lambda_permission" "allow_events_warmup" {
  statement_id  = "AllowExecutionFromEventBridgeWarmup"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.app.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.warmup.arn
}

output "ecr_repository_url" {
  value = aws_ecr_repository.repo.repository_url
}

output "github_actions_role_arn" {
  value = aws_iam_role.github_actions_ecr_push.arn
}

output "lambda_function_name" {
  value = aws_lambda_function.app.function_name
}

output "lambda_function_url" {
  value = aws_lambda_function_url.app_url.function_url
}
