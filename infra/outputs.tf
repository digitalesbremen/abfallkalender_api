##
## outputs.tf — stack outputs
##
## These values are printed after apply and can be consumed by scripts/CI.

output "ecr_repository_url" {
  description = "Full ECR repository URL."
  value       = aws_ecr_repository.repo.repository_url
}

output "github_actions_role_arn" {
  description = "ARN of the IAM role assumed by GitHub Actions for ECR pushes."
  value       = aws_iam_role.github_actions_ecr_push.arn
}

output "lambda_function_name" {
  description = "Name of the deployed Lambda function."
  value       = aws_lambda_function.app.function_name
}

output "lambda_function_url" {
  description = "Public Function URL for invoking the Lambda (no API Gateway)."
  value       = aws_lambda_function_url.app_url.function_url
}
