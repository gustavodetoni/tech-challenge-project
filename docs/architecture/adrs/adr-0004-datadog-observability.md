# ADR 0004: Observabilidade Com Datadog

Status: Aceita

## Contexto

A operacao precisa detectar gargalos e falhas em tempo real.
O desafio exige metricas, logs estruturados, correlacao entre requisicoes, healthchecks, dashboards e alertas.

## Decisao

Usar Datadog como solucao principal de observabilidade.
Instalar o agente no EKS via Helm values mantidos no repo `tech-challenge-infra-k8s`.
A API principal emite logs JSON com `correlation_id` por request.

## Consequencias

- A coleta de logs e metricas fica centralizada.
- A API precisa preservar `X-Correlation-ID` recebido ou gerar um novo ID.
- Dashboards e alertas devem usar tags de ambiente, servico e versao.
- A demonstracao depende da configuracao de `DATADOG_API_KEY` e da aplicacao em execucao.

## Relacao Com RFC

- [RFC 0004: Observabilidade com Datadog](../rfcs/rfc-0004-datadog-observability.md)

