
variable "region" {
  description = "AWS region"
  type        = string
  default     = "eu-central-1"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  description = "Public subnet CIDRs"
  type        = list(string)
  default     = ["10.0.1.0/24"]
}

variable "private_subnet_cidrs" {
  description = "Private subnet CIDRs"
  type        = list(string)
  default     = ["10.0.2.0/24"]
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "appdb"
}


#### Secrets

resource "random_password" "db_username" {
  length  = 16
  special = false
  numeric = false
}

resource "random_password" "db_password" {
  length  = 20
  special = true
}

resource "aws_secretsmanager_secret" "rds_master" {
  name = "cg-tracker/rds/master"
}

resource "aws_secretsmanager_secret_version" "rds_master" {
  secret_id     = aws_secretsmanager_secret.rds_master.id
  secret_string = jsonencode({
    username = random_password.db_username.result
    password = random_password.db_password.result
  })
}


resource "aws_secretsmanager_secret_rotation" "rds_master_rotation" {
  secret_id = aws_secretsmanager_secret.rds_master.id
  rotation_lambda_arn = aws_lambda_function.rds_secret_rotation.arn
  rotation_rules {
    automatically_after_days = 30
  }
}


data "aws_secretsmanager_secret" "rds_master" {
  name = aws_secretsmanager_secret.rds_master.name
}

data "aws_secretsmanager_secret_version" "rds_master" {
  secret_id = data.aws_secretsmanager_secret.rds_master.id
}

locals {
  rds_secret = jsondecode(data.aws_secretsmanager_secret_version.rds_master.secret_string)
}
