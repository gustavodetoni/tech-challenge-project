# Fase 4 — Microsserviços e Saga

## Serviços

| Serviço | Responsabilidade | Banco |
|---|---|---|
| OS Service | ciclo de vida da OS, estado da Saga e Outbox | PostgreSQL exclusivo |
| Billing Service | orçamento, cobrança, Mercado Pago e webhook | PostgreSQL exclusivo |
| Execution Service | fila, diagnóstico e execução | MongoDB exclusivo |

Nenhum serviço acessa o banco de outro serviço. As integrações são feitas por REST e RabbitMQ.

## Saga orquestrada

O OS Service inicia a Saga ao publicar `ServiceOrderCreated`. O Billing Service cria o orçamento e publica `PaymentConfirmed` após o webhook do Mercado Pago. O Execution Service consome esse evento e publica `ExecutionStarted`/`ServiceOrderExecutionCompleted`. Em falha, o Execution Service publica `ExecutionFailed`, o OS Service cancela a ordem e publica `PaymentRefundRequested`; o Billing Service executa o estorno e marca o orçamento como `REFUNDED`.

Cada serviço possui Outbox ou persistência local antes da publicação, consumidores idempotentes e filas duráveis. A compensação da criação da OS é `POST /service-orders/{id}/cancel`; falhas de pagamento ou execução permanecem rastreáveis por `saga_id` e podem ser reprocessadas pela fila.

## Operação local

```bash
docker compose -f docker-compose.phase4.yml up --build
```

RabbitMQ Management: `http://localhost:15672` (`guest/guest`). Mercado Pago é simulado quando `MERCADO_PAGO_ACCESS_TOKEN` não está configurado; em homologação deve ser configurado o token de teste.

## Testes

Os testes unitários ficam na pasta `tests/` de cada serviço. O OS Service possui também um cenário BDD executável com Godog em `tests/order_saga_bdd_test.go`, cobrindo o fluxo de pagamento e a compensação após falha de execução. O CI executa `go test`, `go vet`, cobertura mínima de 80%, build e análise Sonar.
