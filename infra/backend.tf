
terraform {
  backend "s3" {
    bucket = "cg-tf-state-2025"
    key    = "terraform.tfstate"
    region = "eu-north-1"
  }
}
