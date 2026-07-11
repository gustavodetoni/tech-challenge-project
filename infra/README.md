# Infraestrutura

Os scripts de infraestrutura ficam em [infra/aws](aws).

## AWS

A infraestrutura AWS e provisionada com Terraform e executada principalmente pela esteira manual `Deploy AWS` no GitHub Actions.

Recursos criados:

- VPC.
- Subnets publicas e privadas.
- NAT Gateway.
- Cluster EKS.
- Node group gerenciado.
- Banco PostgreSQL RDS privado.
- Security Group para acesso ao banco pelos nodes do EKS.

Documentacao detalhada:

- [infra/aws/README.md](aws/README.md)

## Fluxo Recomendado

1. Fazer push na `main`.
2. Aguardar `Quality`, `Sonar` e `Release`.
3. Executar manualmente o workflow `Deploy AWS` com `action=apply`.
4. Ao final da demonstracao, executar o mesmo workflow com `action=destroy`.

O state do Terraform e armazenado em S3 usando o secret `TF_STATE_BUCKET`.
