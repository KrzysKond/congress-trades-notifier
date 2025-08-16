
resource "aws_security_group" "lambda" {
  name        = "lambda-sg"
  description = "Allow Lambda to connect to RDS"
  vpc_id      = aws_vpc.main.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_lambda_function" "load_data" {
  function_name = "load-data-lambda"
  role          = aws_iam_role.lambda_exec.arn
  runtime       = "provided.al2023"
  handler       = "lambda_function.lambda_handler"
  filename      = "build/load_data.zip"

  vpc_config {
    subnet_ids         = aws_subnet.public[*].id
    security_group_ids = [aws_security_group.lambda.id]
  }
}

resource "aws_lambda_function" "transform_data" {
  function_name = "transform-data-lambda"
  role          = aws_iam_role.lambda_exec.arn
  runtime       = "provided.al2023"
  handler       = "lambda_function.lambda_handler"
  filename      = "build/transform_data.zip"

  vpc_config {
    subnet_ids         = aws_subnet.private[*].id
    security_group_ids = [aws_security_group.lambda.id]
  }
}
