# Infra AWS

Este Terraform e legado da fase anterior e foi mantido apenas como referencia historica.
Na entrega atual do Tech Challenge, a infraestrutura oficial esta segregada em:

```text
tech-challenge-infra-k8s
tech-challenge-infra-database
```

Nao use este diretorio como fonte oficial para homologacao/producao nesta fase.

## Recursos Do Fluxo Legado

- VPC com subnets publicas e privadas.
- NAT Gateway para os nodes privados acessarem a internet.
- Cluster Kubernetes EKS.
- Node group gerenciado com instancias `t3.small`.
- Banco PostgreSQL RDS privado.
- Security Group permitindo acesso ao RDS apenas pelos nodes do EKS.

## State Do Terraform

No fluxo legado, a esteira usava backend S3 para manter o state entre execucoes do GitHub Actions.
No fluxo atual, cada repositorio de infraestrutura possui seu proprio backend/state.

Cadastre no GitHub Secrets um nome de bucket globalmente unico:

```text
TF_STATE_BUCKET=tech-challenge-terraform-state-gustavo
```

Na primeira execucao, a propria esteira cria esse bucket se ele ainda nao existir.

## Secrets Da Esteira

Cadastre no GitHub:

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
```

`BREVO_API_KEY` e `BREVO_SENDER_EMAIL` podem ficar vazios se o envio de e-mail nao for demonstrado.

## Aplicar Pela Esteira

Este fluxo nao deve ser usado como deploy oficial da entrega atual.
Ele fica documentado apenas para consulta.

No fluxo legado, ao fazer push na `main`, a cadeia automatica era:

```text
Quality -> Sonar -> Release
```

O deploy em AWS era manual pelo GitHub Actions:

```text
Deploy AWS -> Run workflow -> action=apply
```

Antes de aplicar a infraestrutura, o workflow valida:

- O workflow foi executado a partir da branch `main`.
- A versao do `version.go` possui tag `vX.Y.Z`.
- A tag aponta para o commit atual da `main`.
- `Quality`, `Sonar` e `Release` passaram com sucesso para esse commit.
- A imagem Docker versionada existe no Docker Hub.

Depois das validacoes, o workflow legado executava:

1. `terraform init`.
2. `terraform plan`.
3. `terraform apply`.
4. Usa a imagem Docker ja publicada pela `Release` no Docker Hub.
5. Deploy do banco: provisionamento do RDS via Terraform e execucao de migrations/seed via `Job` no Kubernetes.
6. Aplicacao dos manifests Kubernetes no EKS.
7. Rolling update dos pods da API.
8. Impressao da URL publica da API no log e no resumo do GitHub Actions.

O build da aplicacao, testes automatizados, analise Sonar e build/push da imagem Docker ficam nos workflows anteriores. O deploy e o ultimo workflow da cadeia e somente consome a imagem ja publicada.

## Uso Local Opcional

Para testar localmente:

```bash
cd infra/aws
cp terraform.tfvars.example terraform.tfvars
terraform init \
  -backend-config="bucket=SEU_BUCKET_DE_STATE" \
  -backend-config="key=tech-challenge/aws/terraform.tfstate" \
  -backend-config="region=us-east-1"
terraform plan
terraform apply
```

## Derrubar A Infra

Pelo GitHub Actions, execute manualmente o workflow `Deploy AWS` com a opcao:

```text
action = destroy
```

Localmente:

```bash
cd infra/aws
terraform destroy
```

Depois da demonstracao, derrube a infra para evitar custo de EKS, RDS e NAT Gateway.
