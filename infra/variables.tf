variable "region" {
  type    = string
  default = "eu-west-1"
}

variable "project" {
  type    = string
  default = "taskboard-cloud"
}

variable "db_name" {
  type    = string
  default = "taskboard"
}

variable "db_user" {
  type    = string
  default = "taskboard"
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "jwt_secret" {
  type      = string
  sensitive = true
}
