# Tech Challenge API

API em Go para gestao de oficina mecanica, cobrindo clientes, veiculos, servicos, pecas, usuarios e ordens de servico. A fase 2 evolui a aplicacao para aumentar qualidade, resiliencia, escalabilidade e automacao de infraestrutura/deploy.

## Repositorios Da Entrega

- Aplicacao principal: https://github.com/gustavodetoni/tech-challenge-project
- Lambda Auth CPF/CNPJ: https://github.com/gustavodetoni/tech-challenge-auth-lambda
- Infra Kubernetes: https://github.com/gustavodetoni/tech-challenge-infra-k8s
- Infra Database: https://github.com/gustavodetoni/tech-challenge-infra-database

## Objetivos Da Fase 2

- Manter a aplicacao organizada em camadas, seguindo Clean Architecture/Arquitetura Hexagonal.
- Automatizar qualidade, release e deploy com GitHub Actions.
- Executar a aplicacao em containers Docker.
- Orquestrar a aplicacao em Kubernetes com HPA.
- Provisionar infraestrutura AWS com Terraform.
- Usar banco PostgreSQL gerenciado no deploy cloud.
- Demonstrar escalabilidade automatica e deploy versionado.

## Arquitetura Da Aplicacao

O projeto segue uma organizacao em camadas:

- `internal/domain`: entidades e regras centrais do dominio.
- `internal/application`: casos de uso e orquestracao das regras.
- `internal/interfaces`: adaptadores de entrada, como HTTP, DTOs, mappers e middlewares.
- `internal/infra`: integracoes concretas, como banco PostgreSQL, repositorios, JWT e e-mail.
- `pkg`: utilitarios reutilizaveis e sem dependencia de `internal`.

Fluxo simplificado:

```text
HTTP/Gin -> Controllers -> Application Services -> Domain -> Repository Ports -> Infra/PostgreSQL
```

## Principais APIs

- Autenticacao e usuario autenticado.
- CRUD administrativo de clientes, veiculos, servicos e pecas.
- Controle de estoque de pecas.
- Abertura de ordem de servico.
- Consulta de ordem e status.
- Fluxo de diagnostico, orcamento, aprovacao/rejeicao, execucao, finalizacao e entrega.
- Rotas de cliente protegidas por JWT emitido pela Lambda de autenticacao CPF/CNPJ.
- Endpoint externo para decisao de orcamento.
- Metricas administrativas.

## Documentacao Das APIs

- Swagger local: http://localhost:8080/swagger/index.html
- Swagger publicado: https://hwq42fgalh.execute-api.us-east-1.amazonaws.com/swagger/index.html
- Swagger JSON: [docs/swagger.json](docs/swagger.json)
- Swagger YAML: [docs/swagger.yaml](docs/swagger.yaml)
- Collection Postman: [docs/collections/tech-challenge.postman_collection.json](docs/collections/tech-challenge.postman_collection.json)
- Collection Postman no GitHub: https://github.com/gustavodetoni/tech-challenge-project/blob/main/docs/collections/tech-challenge.postman_collection.json

## Governanca Dos Repositorios

Os quatro repositorios da entrega possuem branch protection ativa na branch `main`, exigindo Pull Request para merge, execucao das validacoes de CI e impedindo commits diretos como fluxo oficial de desenvolvimento.

Branches de homologacao e producao sao atendidas por GitHub Actions. O deploy e automatizado pela esteira: apos o disparo definido no workflow, a pipeline executa validacoes, aplica configuracoes, publica artefatos e realiza o rollout no ambiente AWS sem comandos manuais nos servidores.

## Execucao Local Com Docker Compose

Pre-requisitos:

- Docker
- Docker Compose

Subir API, banco e seed:

```bash
docker compose up --build
```

A API fica disponivel em:

```text
http://localhost:8080
```

Healthcheck:

```bash
curl http://localhost:8080/health
```

Swagger:

```text
http://localhost:8080/swagger/index.html
```

Credenciais iniciais criadas pelo seed:

```text
Email: admin@tech.local
Senha: Senha@123
```

Parar o ambiente:

```bash
docker compose down
```

Remover volumes locais do banco:

```bash
docker compose down -v
```

## Docker

Arquivos principais:

- [Dockerfile](Dockerfile)
- [docker-compose.yml](docker-compose.yml)

O `Dockerfile` gera dois binarios:

- `/api`: aplicacao HTTP.
- `/seed`: aplica migrations e popula dados iniciais quando necessario.

## Kubernetes

Os manifestos mantidos em [k8s](k8s) sao legado/referencia local da aplicacao principal.
Para a entrega atual do Tech Challenge, a fonte oficial de Kubernetes cloud passou a ser o repositorio separado `tech-challenge-infra-k8s`.

- `k8s/base`: manifests base.
- `k8s/overlays/local`: execucao local com Postgres dentro do cluster.
- `k8s/overlays/aws`: referencia legada do deploy AWS antes da segregacao em repositorios.

Recursos contemplados:

- Namespace.
- Deployment da API.
- Service da API.
- ConfigMap.
- Secrets via Kustomize.
- HPA por CPU/memoria.
- ServiceAccount.
- Postgres local via StatefulSet no overlay local.

Aplicar localmente:

```bash
cp k8s/overlays/local/.env.k8s.local.example k8s/overlays/local/.env.k8s.local
kubectl apply -k k8s/overlays/local
```

Acompanhar recursos:

```bash
kubectl get all -n tech-challenge
kubectl get hpa -n tech-challenge
kubectl get pods -n tech-challenge -w
```

Acessar localmente:

```bash
kubectl port-forward -n tech-challenge svc/tech-challenge-api 8080:8080
```

Limpar:

```bash
kubectl delete -k k8s/overlays/local
```

Mais detalhes: [k8s/README.md](k8s/README.md)

Fonte oficial para homologacao/producao:

```text
tech-challenge-infra-k8s
```

## Infraestrutura AWS Com Terraform

Os scripts Terraform mantidos em [infra/aws](infra/aws) sao legado da fase anterior e referencia historica do deploy monorepo.
Para a entrega atual, a infraestrutura foi segregada nos repositorios oficiais:

```text
tech-challenge-infra-k8s
tech-challenge-infra-database
```

Recursos agora provisionados pelos repositorios oficiais:

- VPC.
- Subnets publicas e privadas.
- NAT Gateway.
- Cluster EKS.
- Node group gerenciado.
- RDS PostgreSQL privado.
- Security Group permitindo acesso ao RDS pelos nodes do EKS.
- API Gateway.
- Observabilidade Datadog.

As pastas Terraform oficiais da entrega sao:

- `tech-challenge-infra-k8s/terraform`: VPC, EKS, API Gateway, VPC Link, security groups e integracoes.
- `tech-challenge-infra-database/terraform`: RDS PostgreSQL gerenciado.
- `tech-challenge-auth-lambda/terraform`: Lambda serverless de autenticacao por CPF/CNPJ.

O deploy cloud final e executado pelas esteiras de GitHub Actions usando Terraform e deve seguir a ordem documentada no repo `tech-challenge-infra-k8s`.

Mais detalhes: [infra/aws/README.md](infra/aws/README.md)

## CI/CD

Workflows em [.github/workflows](.github/workflows):

```text
Pull Request -> Quality/Sonar
main/homolog/prod -> Release -> Deploy automatizado da aplicacao
```

A pipeline de deploy da aplicacao usa GitHub Actions para conectar no EKS da AWS, validar o Deployment criado pela infraestrutura Kubernetes, executar migrations em `Job` Kubernetes, atualizar a imagem Docker publicada no Docker Hub e aguardar o rollout do Deployment.

## Observabilidade

A API emite logs estruturados em JSON para cada requisicao HTTP.
Cada request recebe um `X-Correlation-ID`; quando o cliente envia esse header, o valor e preservado, caso contrario a API gera um UUID.
Esse identificador funciona como `traceId`/`correlation_id` da requisicao e permite acompanhar o mesmo fluxo entre API Gateway, API principal, logs, erros padronizados e traces coletados pelo Datadog Agent.

Campos principais dos logs:

- `correlation_id`
- `method`
- `path`
- `route`
- `status`
- `latency_ms`
- `client_ip`
- `user_agent`

O mesmo `correlation_id` e propagado para o contexto da request, header da resposta e respostas JSON de erro padronizadas.
A observabilidade da aplicacao em Kubernetes foi implementada com Datadog Agent instalado no cluster pelo repositorio `tech-challenge-infra-k8s`, coletando metricas de pods/deployments/nodes, logs estruturados, latencia, erros HTTP, healthchecks e traces correlacionados pelo `traceId`.

### Quality

Executa:

- Validacao de versao em Pull Requests.
- Lint.
- Testes automatizados.
- Cobertura minima.

### Sonar

Executa:

- Testes com cobertura.
- Analise SonarQube.

### Release

Executa apos `Sonar` passar na `main`:

- Le a versao em `version.go`.
- Valida se a tag ainda nao existe.
- Build da imagem Docker.
- Push no Docker Hub:

```text
docker.io/DOCKERHUB_USERNAME/tech-challenge-project:X.Y.Z
docker.io/DOCKERHUB_USERNAME/tech-challenge-project:latest
```

- Cria GitHub Release.

### Deploy Da Aplicacao

O workflow `Deploy` automatiza o rollout da aplicacao no EKS usando GitHub Actions.
Ele nao provisiona VPC, EKS, API Gateway ou RDS; essa responsabilidade fica nos repositorios segregados de infraestrutura.

Fluxo previsto:

```text
pull_request -> Quality/Sonar
main/homolog/prod -> Release da imagem Docker
Deploy           -> action=apply, environment=homolog/prod, image_tag=latest
Destroy          -> action=destroy, environment=homolog/prod
```

Inputs do workflow:

```text
action          apply ou destroy
environment     homolog ou prod
image_tag       tag Docker publicada pelo workflow Release
cluster_name    opcional; se vazio usa tech-challenge-<ambiente>-eks
run_migrations  yes para executar /seed antes do rollout
```

O deploy executa:

- Conexao no EKS via AWS Academy.
- Validacao do Deployment criado pelo repo `tech-challenge-infra-k8s`.
- Migrations via `Job` Kubernetes usando o binario `/seed`.
- Atualizacao da imagem Docker no Deployment.
- Aguardo do rollout.

Para remover apenas a aplicacao:

```text
Deploy -> Run workflow -> action=destroy
```

O destroy remove Deployment, HPA, Service, ConfigMap, Secret, ServiceAccount e o Job de migrations da aplicacao.
A infraestrutura cloud deve ser destruida depois pelos repositorios `tech-challenge-auth-lambda`, `tech-challenge-infra-database` e `tech-challenge-infra-k8s`.

## Secrets Do GitHub Actions

Configure:

```text
AWS_ACCESS_KEY_ID      deploy automatizado no EKS
AWS_SECRET_ACCESS_KEY  deploy automatizado no EKS
AWS_SESSION_TOKEN      deploy automatizado no EKS usando AWS Academy
DOCKERHUB_USERNAME     release da imagem Docker
DOCKERHUB_TOKEN        release da imagem Docker
SONAR_TOKEN            analise Sonar
```

Secrets como `DATABASE_URL`, `JWT_SECRET`, `BREVO_API_KEY` e `BREVO_SENDER_EMAIL` sao aplicados no Kubernetes pelo repositorio `tech-challenge-infra-k8s`.

## Teste De Escalabilidade

O projeto possui script k6 em [scripts/k6s](scripts/k6s) para gerar carga e observar o HPA.

Exemplo:

```bash
BASE_URL=http://URL_DA_API REGISTER_VUS=80 HEALTH_RPS=60 DURATION=10m k6 run scripts/k6s/load-test.js
```

Durante o teste:

```bash
kubectl get hpa -n tech-challenge -w
kubectl get pods -n tech-challenge -w
kubectl top pods -n tech-challenge
```

Mais detalhes: [scripts/k6s/README.md](scripts/k6s/README.md)

## Links

- Repositorio: https://github.com/gustavodetoni/tech-challenge-project
- Swagger local: http://localhost:8080/swagger/index.html
- Swagger YAML: https://github.com/gustavodetoni/tech-challenge-project/blob/main/docs/swagger.yaml
- Swagger JSON: https://github.com/gustavodetoni/tech-challenge-project/blob/main/docs/swagger.json
- Postman: https://github.com/gustavodetoni/tech-challenge-project/blob/main/docs/collections/tech-challenge.postman_collection.json
- Deploy homologacao: https://hwq42fgalh.execute-api.us-east-1.amazonaws.com
- Healthcheck homologacao: https://hwq42fgalh.execute-api.us-east-1.amazonaws.com/health
- Deploy producao: mesmo fluxo automatizado de deploy, usando `environment=prod`.

## Documentos Complementares

- Arquitetura global, RFCs, ADRs e sequencias: [docs/architecture](docs/architecture)
- Documentacao C4/Event Storming: [docs/documentation](docs/documentation)
- Relatorios Sonar: [docs/sonar](docs/sonar)
