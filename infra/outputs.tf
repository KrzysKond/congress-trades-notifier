output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.main_lambda.arn
}

output "s3_bucket_name" {
  description = "Name of the S3 bucket"
  value       = aws_s3_bucket.main.bucket
}

output "s3_bucket_arn" {
  description = "ARN of the S3 bucket"
  value       = aws_s3_bucket.main.arn
}

output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "ses_identity" {
  description = "SES email identity"
  value       = aws_ses_email_identity.identity.email
}