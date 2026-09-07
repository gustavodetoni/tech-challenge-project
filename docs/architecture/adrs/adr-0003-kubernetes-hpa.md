# ADR 0003: Autoscaling Kubernetes Com HPA

Status: Aceita

## Contexto

A aplicacao principal precisa escalar conforme aumento de clientes e unidades da oficina.
O desafio exige cluster Kubernetes com escalabilidade e monitoramento de CPU/memoria.

## Decisao

Executar a API principal no EKS usando Deployment, Service e HorizontalPodAutoscaler.
O HPA usa metricas de CPU e memoria para ajustar replicas.

## Configuracao Base

- Deployment `tech-challenge-api`.
- Service interno para exposicao ao API Gateway.
- HPA com minimo de 2 replicas.
- Limites e requests de CPU/memoria definidos no container.
- Probes de startup, readiness e liveness usando `/health`.

## Consequencias

- A API pode absorver variacao de carga sem deploy manual.
- O cluster precisa de metrics-server ou equivalente funcionando.
- Requests/limits devem ser revisados com dados reais de carga.
- Dashboards precisam acompanhar replicas, CPU, memoria e restarts.

## Relacao Com RFC

- [RFC 0001: Escolha da AWS](../rfcs/rfc-0001-aws-cloud-provider.md)
- [RFC 0004: Observabilidade com Datadog](../rfcs/rfc-0004-datadog-observability.md)

