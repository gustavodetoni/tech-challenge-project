output "cluster_name" {
  description = "Nome do cluster EKS."
  value       = module.eks.cluster_name
}

output "database_host" {
  description = "Endpoint do RDS PostgreSQL."
  value       = aws_db_instance.postgres.address
}

output "database_name" {
  description = "Nome do banco criado no RDS."
  value       = aws_db_instance.postgres.db_name
}

output "database_port" {
  description = "Porta do PostgreSQL."
  value       = aws_db_instance.postgres.port
}
