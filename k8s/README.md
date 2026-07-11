# Guia Local Do Kubernetes

Este diretório contém os manifestos Kubernetes do projeto Tech Challenge.

O overlay local usa a imagem publicada no Docker Hub pela esteira de release:

```text
docker.io/gustavodetoni/tech-challenge-project:1.1.3
```

## Requisitos

- Docker Desktop com Kubernetes habilitado.
- `kubectl` configurado para o cluster do Docker Desktop.
- A tag da imagem precisa existir no Docker Hub antes de aplicar os manifestos.

## Validar O Cluster

```bash
kubectl cluster-info
kubectl get nodes
```

## Validar A Imagem Docker

```bash
docker manifest inspect docker.io/gustavodetoni/tech-challenge-project:1.1.3
```

## Configurar Variáveis E Secrets

O Kubernetes não lê o `.env` da aplicação automaticamente. Para o ambiente local, este projeto usa um arquivo específico:

```text
k8s/overlays/local/.env.k8s.local
```

Esse arquivo é ignorado pelo Git e é usado pelo Kustomize para gerar os `Secrets`:

```text
tech-challenge-api-secret
tech-challenge-postgres-secret
```

Crie o arquivo a partir do exemplo:

```bash
cp k8s/overlays/local/.env.k8s.local.example k8s/overlays/local/.env.k8s.local
```

Depois edite:

```bash
k8s/overlays/local/.env.k8s.local
```

Pontos importantes:

- Não use `localhost` no `DATABASE_URL`, porque dentro do Kubernetes o banco é acessado pelo Service `tech-challenge-postgres`.
- Não coloque comentários na mesma linha dos valores, por exemplo `JWT_EXPIRY_MINUTES=10080 # comentario`, porque isso vira parte do valor.
- Não commite o `.env.k8s.local`.

Exemplo correto para o banco local no Kubernetes:

```env
DATABASE_URL=postgres://tech_challenge:tech_challenge@tech-challenge-postgres:5432/tech_challenge?sslmode=disable
```

## Renderizar Os Manifestos

```bash
kubectl kustomize k8s/overlays/local
```

## Simular A Aplicação Dos Manifestos

```bash
kubectl apply --dry-run=client -k k8s/overlays/local
```

## Aplicar Localmente

```bash
kubectl apply -k k8s/overlays/local
```

## Acompanhar Os Recursos

```bash
kubectl get all -n tech-challenge
kubectl get pods -n tech-challenge -w
```

## Verificar Logs

```bash
kubectl logs -n tech-challenge statefulset/tech-challenge-postgres
kubectl logs -n tech-challenge deploy/tech-challenge-api
```

## Acessar A API

```bash
kubectl port-forward -n tech-challenge svc/tech-challenge-api 8080:8080
```

Em outro terminal:

```bash
curl http://localhost:8080/health
```

Resposta esperada:

```json
{
  "status": "ok",
  "version": "1.1.3"
}
```

## Verificar O HPA

```bash
kubectl get hpa -n tech-challenge
kubectl top pods -n tech-challenge
```

Se `kubectl top pods` falhar, instale o metrics-server no cluster local antes de validar o autoscaling.

## Escalar A API Manualmente

Mesmo sem metrics-server, é possível testar múltiplas réplicas da API manualmente.

Subir para 2 pods:

```bash
kubectl scale deployment tech-challenge-api -n tech-challenge --replicas=2
```

Verificar:

```bash
kubectl get pods -n tech-challenge
```

Voltar para 1 pod:

```bash
kubectl scale deployment tech-challenge-api -n tech-challenge --replicas=1
```

## Limpar O Ambiente

```bash
kubectl delete -k k8s/overlays/local
```

Se o volume persistente continuar existindo e você quiser resetar os dados locais do banco:

```bash
kubectl delete pvc -n tech-challenge -l app.kubernetes.io/name=tech-challenge-postgres
```

## Secrets E Segurança

O arquivo `.env.k8s.local.example` contém apenas um modelo. O arquivo real `.env.k8s.local` deve existir somente na sua máquina.

Para AWS ou ambientes compartilhados, forneça secrets por um mecanismo controlado, como secrets Kubernetes gerenciados pelo Terraform, AWS Secrets Manager, External Secrets ou outro fluxo de gerenciamento de secrets.
