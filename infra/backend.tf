
terraform {
  backend "s3" {
    bucket = "cg-fillings"
    key    = "terraform.tfstate"
    region = "eu-north-1"
  }
}
