# Fase 4 — Microsserviços e Saga

## Serviços

| Serviço | Responsabilidade | Banco |
|---|---|---|
| OS Service | ciclo de vida da OS, estado da Saga e Outbox | PostgreSQL exclusivo |
| Billing Service | orçamento, cobrança, Mercado Pago e webhook | PostgreSQL exclusivo |
| Execution Service | fila, diagnóstico e execução | MongoDB exclusivo |

Nenhum serviço acessa o banco de outro serviço. As integrações são feitas por REST e RabbitMQ.

## Saga orquestrada

O OS Service inicia a Saga ao publicar `ServiceOrderCreated`. O Billing Service cria o orçamento e publica `PaymentConfirmed` após o webhook do Mercado Pago. O Execution Service consome esse evento e publica `ExecutionStarted`/`ServiceOrderExecutionCompleted`. O OS Service consome os eventos e atualiza o estado local.

Cada serviço possui Outbox ou persistência local antes da publicação, consumidores idempotentes e filas duráveis. A compensação da criação da OS é `POST /service-orders/{id}/cancel`; falhas de pagamento ou execução permanecem rastreáveis por `saga_id` e podem ser reprocessadas pela fila.

## Operação local

```bash
docker compose -f docker-compose.phase4.yml up --build
```

RabbitMQ Management: `http://localhost:15672` (`guest/guest`). Mercado Pago é simulado quando `MERCADO_PAGO_ACCESS_TOKEN` não está configurado; em homologação deve ser configurado o token de teste.

## Testes

Os testes unitários e o cenário BDD completo foram deliberadamente deixados para a próxima etapa, conforme o escopo de implementação solicitado. Os pipelines já executam `go test`, `go vet`, build e análise Sonar, ficando a regra de cobertura de 80% para quando os testes forem adicionados.
