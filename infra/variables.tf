##
## variables.tf — input variables for this stack
##
## This file defines all user-configurable inputs. Keep defaults conservative.

variable "image_tag" {
  type        = string
  description = "Tag of the ECR image for the Lambda function (must already exist in ECR)."
}

variable "lambda_memory_mb" {
  type        = number
  default     = 512
  description = "Allocated memory (MB) for the Lambda function."
}

variable "lambda_timeout_s" {
  type        = number
  default     = 15
  description = "Timeout (seconds) for the Lambda function."
}

# Optional: fixed reserved concurrency to cap parallel invocations/costs.
# Default = null (no reservation) to avoid account quota issues. If you set a
# number > 0, ensure your account still has at least 10 unreserved concurrency
# remaining as required by AWS.
variable "reserved_concurrency" {
  type        = number
  default     = null
  description = "Optional: reserved concurrent executions for the Lambda function (or null for no reservation)."
}
