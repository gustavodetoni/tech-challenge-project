# RFC 0002: Banco PostgreSQL Gerenciado

Status: Aceita

## Contexto

O dominio da oficina e relacional.
Clientes possuem veiculos, veiculos geram ordens de servico, ordens possuem historico de status, orcamentos versionados, servicos, pecas e movimentos de estoque.
O desafio pede melhoria e documentacao da modelagem, garantindo consistencia e performance.

## Decisao

Usar PostgreSQL gerenciado em Amazon RDS.

## Justificativa

- A aplicacao principal ja usa PostgreSQL.
- O fluxo de abertura, orcamento, aprovacao e baixa de estoque depende de transacoes ACID.
- O modelo usa chaves estrangeiras, constraints, indices e enums, recursos bem suportados pelo PostgreSQL.
- RDS oferece backup gerenciado, criptografia em repouso, logs exportados para CloudWatch e janela de manutencao.

## Ajustes Operacionais

O repositorio `tech-challenge-infra-database` define:

- RDS privado em subnets privadas.
- Security group permitindo acesso apenas de EKS e Lambda Auth.
- Parameter group PostgreSQL.
- Backups diarios.
- Janela de manutencao.
- Logs PostgreSQL exportados.
- Outputs para composicao de `DATABASE_URL` da API e da Lambda.

## Alternativas Consideradas

- MySQL: atende parte do modelo relacional, mas traria migracao desnecessaria.
- SQL Server: mais pesado para a proposta e sem ganho claro para o dominio.
- Banco self-managed no Kubernetes: reduziria uso de servico gerenciado, mas iria contra o requisito de banco gerenciado e aumentaria risco operacional.

## Consequencias

- A aplicacao e a Lambda precisam acessar o banco por rede privada.
- O segredo de senha deve ficar no GitHub Secrets ou secret manager equivalente.
- O state Terraform do banco depende dos outputs do state de Kubernetes/rede.

## Repositorios Impactados

- `tech-challenge-infra-database`
- `tech-challenge-project`
- `tech-challenge-auth-lambda`

