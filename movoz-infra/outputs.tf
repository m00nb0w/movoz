output "db_endpoint" {
  description = "RDS instance endpoint"
  value       = aws_db_instance.movoz_db.endpoint
}

output "db_port" {
  description = "RDS instance port"
  value       = aws_db_instance.movoz_db.port
}

output "db_name" {
  description = "Database name"
  value       = aws_db_instance.movoz_db.db_name
}

output "db_username" {
  description = "Database master username"
  value       = aws_db_instance.movoz_db.username
  sensitive   = true
}

output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.movoz_vpc.id
}

output "subnet_ids" {
  description = "Private subnet IDs"
  value       = aws_subnet.private[*].id
}

output "security_group_id" {
  description = "Database security group ID"
  value       = aws_security_group.db_sg.id
}
