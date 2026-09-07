# ADR 0001: API Gateway Como Entrada Publica

Status: Aceita

## Contexto

A plataforma precisa de uma entrada HTTP unica para rotear autenticacao CPF/CNPJ e chamadas da API principal.
Tambem precisa evitar exposicao direta de componentes internos do Kubernetes.

## Decisao

Usar Amazon API Gateway HTTP API como entrada publica.
O backend da aplicacao principal fica no EKS, acessado por VPC Link e Load Balancer interno.
A Lambda Auth fica exposta pela rota `POST /auth/cpf`.

## Consequencias

- O roteamento publico fica fora do cluster Kubernetes.
- O Load Balancer da API pode permanecer interno.
- Rotas de cliente podem receber authorizer especifico.
- O deploy depende dos outputs da Lambda e do listener do NLB.

## Relacao Com RFC

- [RFC 0005: API Gateway e Rotas Protegidas](../rfcs/rfc-0005-api-gateway-routing.md)

