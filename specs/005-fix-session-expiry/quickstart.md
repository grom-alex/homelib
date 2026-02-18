# Quickstart: Верификация исправления истечения сессии

**Branch**: `005-fix-session-expiry`

## Предварительные условия

- Docker Compose dev-окружение запущено: `docker compose -f docker/docker-compose.dev.yml up -d`
- Зарегистрирован хотя бы один пользователь (через UI или curl)
- Браузер с DevTools открыт для наблюдения за cookies и network

## Шаг 1: Проверка cookie_secure в dev-конфиге

```bash
# Убедиться что cookie_secure: false в dev-конфиге
grep cookie_secure config/config.dev.yaml
# Ожидаемый вывод: cookie_secure: false
```

## Шаг 2: Unit-тесты

```bash
# Backend тесты
cd backend && go test -race -v ./internal/config/...

# Frontend тесты
cd frontend && npx vitest run src/api/__tests__/client.test.ts src/stores/__tests__/auth.test.ts
```

Все тесты должны пройти.

## Шаг 3: Проверка автоматического refresh (FR-001)

1. Открыть `http://localhost/login` в браузере
2. Войти с учётными данными
3. Открыть DevTools → Application → Cookies → localhost
4. Убедиться что cookie `refresh_token` существует (HttpOnly, Path=/)
5. Открыть DevTools → Network
6. **Для быстрой проверки**: Временно установить `access_token_ttl: "30s"` в `config/config.dev.yaml`, пересобрать и подождать 30 секунд
7. Нажать на любой элемент интерфейса (фильтр, пагинация, пункт меню)
8. В Network должен появиться запрос `POST /api/auth/refresh` → 200 OK
9. Исходный запрос должен автоматически повториться и завершиться успешно
10. Никаких ошибок 401 на экране

## Шаг 4: Проверка redirect-back (FR-003, FR-004)

1. Находясь на странице `/authors` или `/books?search=толстой`
2. Подождать истечения access token + refresh token (или вручную удалить cookie `refresh_token` в DevTools)
3. Нажать на любой элемент → должен произойти redirect на `/login?redirect=/authors` (или `/login?redirect=/books?search=толстой`)
4. Войти снова
5. После входа — автоматически перенаправлен на `/authors` (или `/books?search=толстой`)

## Шаг 5: Проверка кнопки "Выход" (FR-006)

1. Войти в систему
2. Подождать истечения access token (или вручную очистить cookie)
3. Нажать на иконку пользователя → "Выйти"
4. Должен произойти корректный logout: очистка данных, переход на страницу входа
5. Меню НЕ должно "ломаться" — страница входа должна отображаться корректно

## Шаг 6: Проверка TTL 2 часа (FR-005)

```bash
# Проверить конфигурацию
grep access_token_ttl config/config.dev.yaml config/config.stage.yaml config/config.prod.yaml
# Ожидаемый вывод: access_token_ttl: "2h" во всех файлах
```

## Шаг 7: Проверка что 5xx НЕ вызывает redirect (FR-007)

1. Войти в систему
2. Остановить контейнер postgres: `docker compose -f docker/docker-compose.dev.yml stop postgres`
3. Нажать на элемент интерфейса
4. Должна появиться ошибка о проблеме с сервером, НЕ redirect на login
5. Запустить postgres обратно: `docker compose -f docker/docker-compose.dev.yml start postgres`

## Шаг 8: Проверка на стейджинге после деплоя

1. Убедиться что `config/config.stage.yaml` содержит `cookie_secure: false`
2. Задеплоить: `./scripts/deploy-stage.sh --tag <TAG>`
3. Открыть `http://10.0.100.21:8081/`
4. Войти, подождать 2+ часов (или временно уменьшить TTL для тестирования)
5. Нажать на элемент → должен произойти прозрачный refresh

## Контрольный список

- [ ] Cookie `refresh_token` сохраняется в браузере после входа
- [ ] Автоматический refresh при истечении access token
- [ ] Redirect на `/login?redirect=<path>` при истечении refresh token
- [ ] Redirect-back на исходную страницу после повторного входа
- [ ] Кнопка "Выход" работает в любом состоянии
- [ ] TTL access token = 2 часа
- [ ] 5xx ошибки не вызывают redirect на login
- [ ] Все unit-тесты проходят
