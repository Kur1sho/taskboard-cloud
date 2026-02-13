terraform {
  backend "s3" {
    bucket         = "taskboard-cloud-tfstate-733090080936"
    key            = "taskboard-cloud/terraform.tfstate"
    region         = "eu-west-1"
    dynamodb_table = "taskboard-cloud-tf-locks"
    encrypt        = true
  }
}
