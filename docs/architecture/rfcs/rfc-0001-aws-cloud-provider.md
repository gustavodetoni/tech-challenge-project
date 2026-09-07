# RFC 0001: Escolha Da AWS Como Provedor Cloud

Status: Aceita

## Contexto

O desafio exige API Gateway, function serverless, banco gerenciado, cluster Kubernetes escalavel, Terraform, CI/CD e observabilidade.
A solucao tambem precisa operar em multiplas unidades da oficina, com seguranca, disponibilidade e automacao.

## Decisao

Usar AWS como provedor cloud da entrega.

Servicos principais:

- Amazon EKS para Kubernetes.
- Amazon API Gateway HTTP API para entrada publica e roteamento.
- AWS Lambda para autenticacao CPF/CNPJ.
- Amazon RDS PostgreSQL para banco gerenciado.
- Amazon VPC, subnets publicas/privadas, security groups e NAT Gateway para rede.
- Amazon CloudWatch Logs para logs do RDS e integracao operacional.

## Justificativa

- Todos os blocos obrigatorios do desafio existem como servicos gerenciados na AWS.
- Terraform possui suporte maduro para AWS.
- EKS permite executar a aplicacao principal com HPA e manifests Kubernetes reais.
- API Gateway integra com Lambda e backend privado via VPC Link.
- RDS reduz operacao manual de backup, logs, patching e criptografia.

## Alternativas Consideradas

- Google Cloud: Cloud Run/GKE/Cloud SQL tambem atenderia, mas exigiria reescrever mais decisoes ja iniciadas em AWS.
- Azure: AKS/API Management/Azure Functions atenderia, mas a equipe ja estruturou Terraform e repositorios em AWS.
- Kubernetes puro com Kong/Traefik: reduziria dependencia de API Gateway AWS, mas aumentaria responsabilidade operacional.

## Consequencias

- A subida depende de secrets AWS no GitHub Actions.
- O custo de EKS, RDS e NAT Gateway precisa ser controlado na demonstracao.
- A ordem de deploy precisa respeitar dependencias entre VPC, banco, Lambda e API Gateway.

## Repositorios Impactados

- `tech-challenge-infra-k8s`
- `tech-challenge-infra-database`
- `tech-challenge-auth-lambda`
- `tech-challenge-project`

