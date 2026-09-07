# Arquitetura Global

Esta pasta concentra a documentacao arquitetural global do Tech Challenge.
Ela complementa os READMEs dos quatro repositorios da entrega:

- `tech-challenge-project`: aplicacao principal em Go.
- `tech-challenge-auth-lambda`: autenticacao serverless por CPF/CNPJ.
- `tech-challenge-infra-k8s`: VPC, EKS, API Gateway, manifests Kubernetes, HPA e observabilidade.
- `tech-challenge-infra-database`: RDS PostgreSQL e modelagem relacional.

## RFCs

- [RFC 0001: Escolha da AWS](rfcs/rfc-0001-aws-cloud-provider.md)
- [RFC 0002: Banco PostgreSQL Gerenciado](rfcs/rfc-0002-managed-postgresql.md)
- [RFC 0003: Autenticacao CPF/CNPJ com JWT](rfcs/rfc-0003-cpf-jwt-authentication.md)
- [RFC 0004: Observabilidade com Datadog](rfcs/rfc-0004-datadog-observability.md)
- [RFC 0005: API Gateway e Rotas Protegidas](rfcs/rfc-0005-api-gateway-routing.md)

## ADRs

- [ADR 0001: API Gateway como entrada publica](adrs/adr-0001-api-gateway-entrypoint.md)
- [ADR 0002: Lambda Auth CPF/CNPJ](adrs/adr-0002-lambda-auth-cpf-jwt.md)
- [ADR 0003: Autoscaling Kubernetes com HPA](adrs/adr-0003-kubernetes-hpa.md)
- [ADR 0004: Observabilidade com Datadog](adrs/adr-0004-datadog-observability.md)

## Diagramas De Sequencia

- [Sequencia de abertura de ordem de servico](sequences/open-service-order.md)

