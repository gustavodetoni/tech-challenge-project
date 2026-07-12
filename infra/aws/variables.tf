variable "project_name" {
  description = "Nome base usado nos recursos criados na AWS."
  type        = string
  default     = "tech-challenge"
}

variable "aws_region" {
  description = "Regiao AWS onde os recursos serao criados."
  type        = string
  default     = "us-east-1"
}

variable "cluster_version" {
  description = "Versao do Kubernetes usada no EKS."
  type        = string
  default     = "1.33"
}

variable "db_name" {
  description = "Nome do banco PostgreSQL."
  type        = string
  default     = "tech_challenge"
}

variable "db_user" {
  description = "Usuario master do PostgreSQL."
  type        = string
  default     = "tech_challenge"
}

variable "db_password" {
  description = "Senha master do PostgreSQL."
  type        = string
  sensitive   = true
}

variable "tags" {
  description = "Tags aplicadas aos recursos."
  type        = map(string)

  default = {
    Project = "tech-challenge"
    Managed = "terraform"
  }
}
