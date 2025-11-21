terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "eu-central-1"
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

output "ecr_repository_url" {
  value = aws_ecr_repository.repo.repository_url
}

output "github_actions_role_arn" {
  value = aws_iam_role.github_actions_ecr_push.arn
}
