# Quickstart: Настройка GitHub CI/CD и GitHub Flow

**Feature Branch**: `001-github-ci-setup`

## Проверка работоспособности CI

### 1. Убедиться, что workflow существует

```bash
ls .github/workflows/ci.yml
```

### 2. Запушить ветку и создать PR

```bash
git push -u origin 001-github-ci-setup
gh pr create --title "Setup GitHub CI/CD" --body "Initial CI pipeline"
```

### 3. Проверить запуск CI

Перейти на вкладку **Actions** в репозитории GitHub. Workflow `CI` должен автоматически запуститься.

### 4. Проверить status check в PR

В интерфейсе PR в секции "Checks" должен отображаться статус `CI` (pass/fail).

### 5. Проверить branch protection

```bash
gh api repos/grom-alex/homelib/rulesets --jq '.[] | {id, name, enforcement}'
```

Должен вернуть ruleset с `name: "Protect master"` и `enforcement: "active"`.

## Проверка edge case: отсутствие Go-кода

Если `go.mod` не существует в корне репозитория, CI workflow должен:
- Запуститься (job не пропускается)
- Все Go-шаги — пропуститься (conditional skip)
- Итоговый статус — **success** (зелёная галочка)

## Ожидаемая структура файлов

```
.github/
└── workflows/
    └── ci.yml          # GitHub Actions workflow

.golangci.yml           # Конфигурация golangci-lint v2
```
