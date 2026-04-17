# Tech Challenge API

API em Go (Gin) com PostgreSQL para gerenciamento de **clientes**, **veículos**, **serviços**, **peças** e **ordens de serviço**, com autenticação via JWT e documentação Swagger.

## Como rodar

**Pré-requisito:** ter **Docker** e **Docker Compose** instalados.

1. Baixe o projeto:

```bash
git clone https://github.com/gustavodetoni/tech-challenge-project.git
cd tech-challenge-project
```

2. Suba a aplicação (API + banco) com 1 comando:

```bash
docker compose up --build
```

Pronto: o banco e a API sobem juntos, e as migrações/initialização do banco são executadas automaticamente na inicialização.

## Swagger (clique para abrir)

http://localhost:8080/swagger/index.html
