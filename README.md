# Tech Challenge API

API em Go para gestao de oficina mecanica, cobrindo clientes, veiculos, servicos, pecas, usuarios e ordens de servico. A fase 2 evolui a aplicacao para aumentar qualidade, resiliencia, escalabilidade e automacao de infraestrutura/deploy.

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
- Swagger JSON: [docs/swagger.json](docs/swagger.json)
- Swagger YAML: [docs/swagger.yaml](docs/swagger.yaml)
- Collection Postman: [docs/collections/tech-challenge.postman_collection.json](docs/collections/tech-challenge.postman_collection.json)

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

Os manifestos ficam em [k8s](k8s):

- `k8s/base`: manifests base.
- `k8s/overlays/local`: execucao local com Postgres dentro do cluster.
- `k8s/overlays/aws`: deploy AWS usando RDS e Service `LoadBalancer`.

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

## Infraestrutura AWS Com Terraform

Os scripts Terraform ficam em [infra/aws](infra/aws).

Recursos provisionados:

- VPC.
- Subnets publicas e privadas.
- NAT Gateway.
- Cluster EKS.
- Node group gerenciado.
- RDS PostgreSQL privado.
- Security Group permitindo acesso ao RDS pelos nodes do EKS.

O deploy principal usa Terraform pela esteira do GitHub Actions. O state fica em bucket S3 configurado pelo secret `TF_STATE_BUCKET`.

Mais detalhes: [infra/aws/README.md](infra/aws/README.md)

## CI/CD

Workflows em [.github/workflows](.github/workflows):

```text
Quality -> Sonar -> Release -> Deploy AWS manual
```

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

### Deploy AWS

Executado manualmente pelo GitHub Actions:

```text
Deploy AWS -> Run workflow -> action=apply
```

Antes do deploy, valida:

- Branch `main`.
- Tag `vX.Y.Z` da versao atual.
- Tag apontando para o commit atual da `main`.
- `Quality`, `Sonar` e `Release` com sucesso para o commit.
- GitHub Release existente.
- Imagem Docker versionada existente no Docker Hub.

Depois executa:

- `terraform init`.
- `terraform plan`.
- `terraform apply`.
- Gera Secret Kubernetes com `DATABASE_URL`.
- Aplica manifests `k8s/overlays/aws`.
- Executa deploy do banco via `Job` Kubernetes usando `/seed`.
- Atualiza imagem do Deployment.
- Aguarda rollout.
- Imprime URL publica da API no log e no resumo do GitHub Actions.

Para derrubar a infra:

```text
Deploy AWS -> Run workflow -> action=destroy
```

O destroy remove os manifests Kubernetes, quando o cluster ainda existe, e depois executa `terraform destroy`.

## Secrets Do GitHub Actions

Configure:

```text
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
DOCKERHUB_USERNAME
DOCKERHUB_TOKEN
TF_STATE_BUCKET
TF_VAR_DB_PASSWORD
JWT_SECRET
BREVO_API_KEY
BREVO_SENDER_EMAIL
SONAR_TOKEN
```

`BREVO_API_KEY` e `BREVO_SENDER_EMAIL` podem ficar vazios caso o envio de e-mail nao seja demonstrado.

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

## Documentos Complementares

- Documentacao C4/Event Storming: [docs/documentation](docs/documentation)
- Relatorios Sonar: [docs/sonar](docs/sonar)
