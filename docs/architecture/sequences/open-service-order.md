# Sequencia: Abertura De Ordem De Servico

Este diagrama descreve o fluxo principal de abertura de ordem de servico pela area administrativa.
O cliente usa o CPF/CNPJ no cadastro da OS, mas a autenticacao administrativa continua por JWT de usuario interno.

```mermaid
sequenceDiagram
    autonumber
    actor User as Usuario Admin/Manager
    participant Gateway as API Gateway
    participant API as API Principal (EKS)
    participant AuthMW as Auth Middleware
    participant Controller as AdminServiceOrderFlowController
    participant UseCase as ServiceOrderFlowUseCase
    participant Repos as Repositories
    participant DB as RDS PostgreSQL
    participant Obs as Datadog/Logs

    User->>Gateway: POST /admin/service-orders\nAuthorization: Bearer <admin_jwt>
    Gateway->>API: Encaminha request para API no EKS
    API->>AuthMW: Valida JWT administrativo
    AuthMW-->>API: Claims do usuario autenticado
    API->>Controller: Executa CreateDraft
    Controller->>UseCase: CreateDraft(input)

    UseCase->>Repos: Buscar/criar cliente por documento
    Repos->>DB: SELECT/INSERT clients
    DB-->>Repos: Cliente

    UseCase->>Repos: Buscar/criar veiculo por placa
    Repos->>DB: SELECT/INSERT vehicles
    DB-->>Repos: Veiculo

    UseCase->>Repos: Validar servicos e pecas
    Repos->>DB: SELECT services/parts
    DB-->>Repos: Itens validos

    UseCase->>Repos: Criar OS, orcamento draft e historico
    Repos->>DB: INSERT service_orders, budgets, budget_items, status_history
    DB-->>Repos: Transacao confirmada

    UseCase-->>Controller: service_order_id, code, budget_id, status, total
    Controller-->>API: 201 Created
    API-->>Gateway: Resposta JSON
    Gateway-->>User: Dados da OS criada
    API->>Obs: Log JSON com correlation_id, rota, status e latencia
```

## Regras Do Fluxo

- A rota exige JWT administrativo com role `ADMIN` ou `MANAGER`.
- O documento do cliente e validado pela aplicacao principal durante a abertura da OS.
- A criacao da OS, do orcamento inicial e do historico deve ocorrer de forma transacional.
- Se algum servico ou peca nao existir, a API responde `400`.
- Se houver conflito de dados, a API responde `409`.
- Toda resposta de erro deve incluir `correlation_id` quando a request passa pelo middleware de observabilidade.

## Endpoint

```text
POST /admin/service-orders
```

## Repositorios Envolvidos

- `tech-challenge-project`: regra de negocio e endpoint HTTP.
- `tech-challenge-infra-k8s`: API Gateway, EKS e manifests da API.
- `tech-challenge-infra-database`: RDS PostgreSQL e modelagem.

