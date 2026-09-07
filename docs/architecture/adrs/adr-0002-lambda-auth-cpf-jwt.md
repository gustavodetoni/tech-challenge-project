# ADR 0002: Lambda Auth CPF/CNPJ

Status: Aceita

## Contexto

Clientes precisam acessar rotas sensiveis sem login/senha tradicional.
O desafio exige function serverless para validar CPF, consultar cliente na base e emitir JWT.

## Decisao

Implementar a autenticacao de cliente em uma Lambda separada.
A Lambda consulta o PostgreSQL privado e emite JWT assinado.
A API principal valida o JWT de cliente nas rotas `/client`.

## Consequencias

- A autenticacao por documento fica isolada da API principal.
- Lambda precisa de acesso privado ao RDS.
- `JWT_SECRET`, issuer e tempo de expiracao devem ser configurados de forma consistente entre Lambda e API.
- As rotas `/client` deixam de depender de documento em query/body para autorizar acesso.

## Relacao Com RFC

- [RFC 0003: Autenticacao CPF/CNPJ com JWT](../rfcs/rfc-0003-cpf-jwt-authentication.md)

