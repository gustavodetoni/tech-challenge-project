# RFC 0004: Observabilidade Com Datadog

Status: Aceita

## Contexto

O desafio exige visibilidade sobre funcionamento do sistema, incluindo:

- Latencia das APIs.
- CPU e memoria do Kubernetes.
- Healthchecks e uptime.
- Alertas para falhas no processamento de ordens de servico.
- Logs estruturados JSON com correlacao entre requisicoes.
- Dashboards de volume diario de OS, tempo medio por status e erros de integracao.

## Decisao

Usar Datadog como ferramenta principal de observabilidade.

## Justificativa

- O agente Datadog possui instalacao padrao via Helm em Kubernetes.
- Coleta metricas de pods, deployments, nodes e containers.
- Permite dashboards e alertas sobre APIs, Kubernetes, logs e integracoes.
- A aplicacao principal agora emite logs JSON com `correlation_id`, status e latencia.

## Escopo Monitorado

- API Gateway: volume, latencia e status HTTP.
- API principal: logs JSON, erros 4xx/5xx, latencia por rota e correlation ID.
- Kubernetes: CPU, memoria, replicas, restarts e health probes.
- Banco: conexoes, logs PostgreSQL e erros de acesso.
- Lambda Auth: falhas de autenticacao e latencia da funcao.
- Ordens de servico: falhas em criacao, revisao, envio, aprovacao, rejeicao, finalizacao e entrega.

## Alternativas Consideradas

- New Relic: atenderia tecnicamente, mas a estrutura criada ja usa Datadog.
- Prometheus/Grafana: boa alternativa open source, mas demandaria mais operacao para logs, alertas e dashboards integrados.
- Apenas CloudWatch: reduz ferramentas, mas fica menos completo para Kubernetes e traces.

## Consequencias

- E necessario configurar `DATADOG_API_KEY`.
- Dashboards e monitores dependem da subida real para validar nomes de metricas e tags.
- Os logs precisam manter formato JSON em producao.

## Repositorios Impactados

- `tech-challenge-infra-k8s`
- `tech-challenge-project`
- `tech-challenge-auth-lambda`
- `tech-challenge-infra-database`

