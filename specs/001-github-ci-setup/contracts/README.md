# Contracts: GitHub CI/CD Setup

Данная фича не создаёт API-эндпоинтов приложения.

Контракты определяются конфигурацией GitHub:

1. **CI Workflow** (`.github/workflows/ci.yml`) — контракт на набор проверок при push/PR
2. **Branch Protection Ruleset** — контракт на обязательные status checks для merge в master
3. **golangci-lint config** (`.golangci.yml`) — контракт на правила линтинга Go-кода

Детали см. в [data-model.md](../data-model.md).
