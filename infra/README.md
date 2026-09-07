# Infraestrutura

Este diretorio foi mantido apenas como legado/referencia da fase anterior, quando a aplicacao principal concentrava tambem a infraestrutura AWS.

Na entrega atual do Tech Challenge, a fonte oficial da infraestrutura cloud foi segregada em repositorios dedicados:

```text
tech-challenge-infra-k8s
tech-challenge-infra-database
```

## AWS Legado

A infraestrutura em [infra/aws](aws) nao deve ser usada como fonte oficial de homologacao/producao nesta fase.
Ela permanece no repositorio apenas para consulta historica e comparacao com a arquitetura anterior.

Os recursos equivalentes agora ficam separados assim:

- `tech-challenge-infra-k8s`: VPC, EKS, API Gateway, manifests Kubernetes, HPA e observabilidade.
- `tech-challenge-infra-database`: RDS PostgreSQL, parameter group, backups, logs, security group e outputs do banco.

Documentacao detalhada:

- [infra/aws/README.md](aws/README.md)

## Fluxo Recomendado

1. Gerar a imagem da aplicacao principal pelo workflow de release deste repositorio.
2. Aplicar `tech-challenge-infra-k8s` para criar rede, EKS, API Gateway e security group da Lambda.
3. Aplicar `tech-challenge-infra-database` para criar o RDS PostgreSQL.
4. Aplicar `tech-challenge-auth-lambda` para publicar a autenticacao CPF/JWT.
5. Reaplicar `tech-challenge-infra-k8s` com os outputs da Lambda e o listener do NLB.

O state oficial do Terraform passa a ser administrado pelos repositorios segregados.
