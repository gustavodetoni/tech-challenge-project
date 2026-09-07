# RFC 0003: Autenticacao CPF/CNPJ Com JWT

Status: Aceita

## Contexto

O desafio exige proteger rotas sensiveis da aplicacao com autenticacao via CPF e criar uma function serverless capaz de:

- Validar CPF/CNPJ.
- Consultar existencia e status do cliente na base.
- Gerar e devolver JWT valido para consumo das APIs protegidas.

## Decisao

Criar uma Lambda dedicada para autenticacao de cliente por CPF/CNPJ e emitir JWT do tipo `CLIENT`.
A aplicacao principal valida esse token nas rotas `/client`.

## Fluxo

1. Cliente chama `POST /auth/cpf` com CPF/CNPJ.
2. API Gateway encaminha a requisicao para a Lambda.
3. Lambda normaliza e valida o documento.
4. Lambda consulta o cliente no PostgreSQL.
5. Lambda verifica se o cliente esta ativo.
6. Lambda gera JWT assinado com `JWT_SECRET`.
7. Cliente consome rotas `/client` com `Authorization: Bearer <token>`.
8. API principal valida issuer, assinatura, validade, tipo `CLIENT` e documento do token.

## Justificativa

- Mantem a autenticacao de cliente desacoplada da aplicacao principal.
- Atende ao requisito serverless.
- Permite proteger apenas rotas sensiveis de cliente.
- Evita expor documento em query string para rotas protegidas.

## Alternativas Consideradas

- Login tradicional de cliente com senha: adicionaria gestao de credenciais nao solicitada pelo escopo.
- Autenticacao apenas no API Gateway: reduziria codigo na API, mas a API ainda precisa conhecer o documento autenticado para validar acesso a OS.
- Token opaco em banco/cache: exigiria estado adicional e nao traz ganho para a demonstracao.

## Consequencias

- `JWT_SECRET` precisa ser compartilhado com seguranca entre Lambda e aplicacao principal.
- A validade do token precisa ser curta o suficiente para reduzir risco em caso de vazamento.
- A API principal deve rejeitar token administrativo nas rotas de cliente.

## Repositorios Impactados

- `tech-challenge-auth-lambda`
- `tech-challenge-project`
- `tech-challenge-infra-k8s`
- `tech-challenge-infra-database`

