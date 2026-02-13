output "alb_dns_name" {
  value = aws_lb.this.dns_name
}

output "ecr_auth" {
  value = aws_ecr_repository.auth.repository_url
}

output "ecr_tasks" {
  value = aws_ecr_repository.tasks.repository_url
}
