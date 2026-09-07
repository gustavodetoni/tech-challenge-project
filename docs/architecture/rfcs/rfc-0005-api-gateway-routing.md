# RFC 0005: API Gateway E Rotas Protegidas

Status: Aceita

## Contexto

O desafio exige API Gateway para controle e roteamento, e protecao das rotas sensiveis da aplicacao via CPF.
A aplicacao possui rotas publicas, rotas administrativas com JWT interno e rotas de cliente com JWT emitido pela Lambda CPF/CNPJ.

## Decisao

Usar Amazon API Gateway HTTP API como entrada publica.
O gateway roteia:

- `POST /auth/cpf` para a Lambda Auth.
- Rotas gerais para a API principal no EKS via VPC Link.
- Rotas sensiveis `/client` com Lambda Authorizer.

Rotas protegidas:

- `GET /client/service-orders/{code}`
- `GET /client/service-orders/{code}/status`
- `POST /client/service-orders/{code}/budget/approve`
- `POST /client/service-orders/{code}/budget/reject`

## Justificativa

- Evita expor diretamente o Load Balancer interno do Kubernetes.
- Centraliza entrada publica e roteamento.
- Permite proteger apenas as rotas sensiveis de cliente, sem bloquear o proxy inteiro.
- Mantem compatibilidade com Lambda Auth e EKS.

## Alternativas Consideradas

- Ingress Controller no EKS: reduz servicos AWS, mas coloca mais responsabilidade no cluster.
- Kong ou Traefik: sao validos, mas aumentam operacao dentro do Kubernetes para esta entrega.
- Proteger tudo no Gateway: simples, porem bloquearia rotas publicas e administrativas que usam fluxos de autenticacao diferentes.

## Consequencias

- A primeira execucao do repo `tech-challenge-infra-k8s` cria rede, EKS e API Gateway.
- Apos publicar Lambda e NLB da API, o repo `tech-challenge-infra-k8s` deve ser reaplicado com os ARNs finais.
- A API principal continua validando JWT internamente para defesa em profundidade.

## Repositorios Impactados

- `tech-challenge-infra-k8s`
- `tech-challenge-auth-lambda`
- `tech-challenge-project`
