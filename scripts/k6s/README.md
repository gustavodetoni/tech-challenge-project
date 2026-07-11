# Teste de Carga com k6

Este diretório contém um script k6 para gerar carga na API e observar o autoscaling dos pods no Kubernetes.

O teste foi pensado para Kubernetes com banco vazio, sem seed. Ele usa endpoints públicos:

- `GET /health`
- `POST /auth/register`

Por padrão, cada iteração faz várias chamadas de healthcheck e cria uma conta nova com email único.

## Pré-requisitos

- API rodando localmente ou exposta via `kubectl port-forward`.
- Banco com as migrações aplicadas.
- k6 instalado.

Instalação no macOS:

```bash
brew install k6
```

## Rodar com Docker Compose

Suba a API:

```bash
docker compose up --build
```

Em outro terminal, rode o teste:

```bash
k6 run scripts/k6s/load-test.js
```

## Rodar com Kubernetes Local

Suba os manifestos:

```bash
kubectl apply -k k8s/overlays/local
```

Exponha a API localmente:

```bash
kubectl port-forward -n tech-challenge svc/tech-challenge-api 8080:8080
```

Em outro terminal, rode o teste:

```bash
k6 run scripts/k6s/load-test.js
```

## Configurar Carga

Variáveis disponíveis:

- `BASE_URL`: URL base da API. Padrão: `http://localhost:8080`.
- `REGISTER_VUS` ou `VUS`: usuários virtuais simultâneos no endpoint de registro. Padrão: `40`.
- `HEALTH_RPS`: requisições por segundo no endpoint `/health`. Padrão: `30`.
- `DURATION`: duração do platô de carga. Padrão: `5m`.
- `RAMP_UP`: tempo para chegar no pico de VUs. Padrão: `10s`.
- `RAMP_DOWN`: tempo para encerrar a carga. Padrão: `20s`.
- `REGISTER_ENABLED`: habilita criação de contas. Padrão: `true`.
- `SLEEP_SECONDS`: pausa entre iterações de cada VU de registro. Padrão: `0.2`.
- `EMAIL_DOMAIN`: domínio usado nos emails gerados. Padrão: `k6.local`.
- `PASSWORD`: senha das contas criadas. Padrão: `Senha@123`.
- `RUN_ID`: prefixo único da execução. Padrão: timestamp atual.

Exemplo mais pesado:

```bash
BASE_URL=http://localhost:8080 REGISTER_VUS=80 HEALTH_RPS=60 DURATION=10m k6 run scripts/k6s/load-test.js
```

Exemplo só com healthcheck, sem criar usuários:

```bash
REGISTER_ENABLED=false HEALTH_RPS=200 DURATION=10m k6 run scripts/k6s/load-test.js
```

Exemplo apontando para um LoadBalancer ou Ingress:

```bash
BASE_URL=http://SEU-ENDERECO-DA-API REGISTER_VUS=80 HEALTH_RPS=60 DURATION=15m k6 run scripts/k6s/load-test.js
```

## Acompanhar Escala dos Pods

Em outro terminal:

```bash
kubectl get hpa -n tech-challenge -w
```

E também:

```bash
kubectl get pods -n tech-challenge -w
```

Para ver consumo de CPU e memória:

```bash
kubectl top pods -n tech-challenge
```

Se `kubectl top pods` não funcionar, instale ou habilite o metrics-server no cluster.

## Critérios do Teste

O script falha se:

- Mais de 1% das requisições falharem.
- O percentil 95 de latência passar de 1000 ms.
- Menos de 99% dos checks passarem.

Esses limites podem ser ajustados no objeto `options` dentro de `load-test.js`.

## Observações

O endpoint `/auth/register` grava dados no banco. Em execuções longas, ele pode criar muitos usuários.

Para reduzir escrita no banco e focar em escalar pods via healthcheck, use:

```bash
REGISTER_ENABLED=false HEALTH_RPS=300 DURATION=10m k6 run scripts/k6s/load-test.js
```

Para evitar conflito de email ao repetir uma execução com o mesmo `RUN_ID`, deixe o padrão por timestamp ou informe um valor novo:

```bash
RUN_ID=$(date +%s) k6 run scripts/k6s/load-test.js
```
