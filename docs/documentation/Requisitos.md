# Requisitos da Aplicação

## Requisitos funcionais (RF)

- Health check: `/health`
- Autenticação: registrar, login, `/me`
- Administração de usuários: atualizar usuários
- Clientes: CRUD + listagem
- Veículos: CRUD + listagem por cliente
- Peças/insumos: CRUD + listagem
- Estoque: ajustar estoque (movimentação)
- Serviços: CRUD + listagem
- OS: criar, listar (status), detalhar
- Orçamento: revisar, enviar para aprovação
- Fluxo OS: iniciar diagnóstico, finalizar, entregar
- Cliente: consultar OS por código + CPF/CNPJ
- Cliente: aprovar/rejeitar orçamento (motivo)
- Métricas: médias por período/serviço
- Docker: build, compose + seed dados

## Requisitos não funcionais (RNF)

- Segurança: JWT (Bearer) + roles
- Validação de entrada (path/query/body)
- Desempenho: paginação e consultas eficientes
- Confiabilidade: transações e integridade (OS/estoque)
- Observabilidade: logs, métricas, `/health` monitorado
- Documentação: Swagger/OpenAPI atualizado
- Manutenibilidade: arquitetura em camadas, testes, lint/sonar
- Operação: Docker/Compose e config por env vars
