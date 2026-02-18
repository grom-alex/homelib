# Research: Исправление истечения сессии

**Branch**: `005-fix-session-expiry` | **Date**: 2026-02-18

## Исследование 1: Корневая причина ошибки 401

### Decision: `cookie_secure: true` в stage-конфиге при HTTP-соединении

### Rationale

**Проблема**: Staging-сервер работает по HTTP (`http://10.0.100.21:8081`), но в `config/config.stage.yaml` установлено `cookie_secure: true`. Флаг `Secure` в cookie означает, что браузер отправляет cookie **ТОЛЬКО** через HTTPS. На HTTP-соединении:

1. Сервер отвечает `Set-Cookie: refresh_token=...; Secure; HttpOnly; SameSite=Lax`
2. Браузер (Chrome, Firefox) **отклоняет** cookie, т.к. соединение не HTTPS
3. Cookie никогда не сохраняется в браузере
4. При попытке refresh → `POST /api/auth/refresh` → cookie не отправляется → backend возвращает `401 "no refresh token"`

**Воспроизведение**:

1. Пользователь входит → получает access_token в JSON-ответе (хранится в памяти) + refresh_token в Set-Cookie (отклоняется браузером)
2. Access token работает 15 минут (TTL по умолчанию)
3. После 15 мин idle → запрос получает 401 → interceptor пытается refresh → нет cookie → 401 → "Request failed with status code 401"

**Файл**: `config/config.stage.yaml:19` — `cookie_secure: true`

### Alternatives Considered

| Вариант | Описание | Отвергнут потому что |
|---------|----------|---------------------|
| Включить HTTPS на стейджинге | Настроить TLS для nginx | Избыточно для внутренней сети, добавляет сложность управления сертификатами |
| `cookie_secure: false` на стейджинге | Отключить Secure flag для HTTP | **ВЫБРАН** — простое и корректное решение для HTTP-окружения |
| Передавать refresh token в body вместо cookie | Полностью изменить архитектуру токенов | Нарушает §4.I конституции (refresh token ТОЛЬКО в httpOnly cookie) |

---

## Исследование 2: Вторичный баг — store не очищается при неудачном refresh

### Decision: Добавить callback для синхронизации interceptor ↔ Pinia store

### Rationale

**Проблема**: Когда interceptor в `client.ts` не может обновить токен (refresh fails), он выполняет:

```typescript
catch (refreshError) {
  accessToken = null                    // ← только module-level переменная
  onRefreshFailed(refreshError)
  router.push({ name: 'login' })       // ← пытается перейти на /login
  return Promise.reject(error)
}
```

Но Pinia store **не очищается** — `auth.isAuthenticated` остаётся `true`. Router guard на странице `/login` видит авторизованного пользователя и перенаправляет обратно на каталог:

```typescript
// router/index.ts
if (to.name === 'login' && auth.isAuthenticated) return { name: 'catalog' }
```

Результат: пользователь никогда не попадает на страницу входа, видит 401 ошибки.

**Прямой импорт store в client.ts невозможен** — это создаст циклическую зависимость:
- `client.ts` → `stores/auth.ts` → `api/client.ts`

### Alternatives Considered

| Вариант | Описание | Результат |
|---------|----------|-----------|
| Callback-паттерн | `client.ts` экспортирует `setOnAuthExpired(callback)`, store регистрирует callback | **ВЫБРАН** — минимальные изменения, нет цикла |
| Event bus | Глобальный событийный механизм | Избыточно для одного события |
| Импорт через dynamic import | `const { useAuthStore } = await import(...)` | Сложнее, потенциальные проблемы с timing |

---

## Исследование 3: Redirect-back после повторного входа

### Decision: Query-параметр `redirect` в URL страницы входа

### Rationale

**Текущее поведение**: `LoginView.vue` после успешного входа всегда перенаправляет на `/books`:

```typescript
await auth.login(loginForm)
router.push('/books')
```

**Нужно**: При перенаправлении на login из-за истечения сессии — сохранить текущий URL и восстановить его после повторного входа.

**Решение**:
1. В `client.ts` при redirect на login — передавать `query: { redirect: currentRoute.fullPath }`
2. В `LoginView.vue` — читать `route.query.redirect` и переходить по нему после login

Стандартный паттерн для Vue Router, используется в vue-router документации.

### Alternatives Considered

| Вариант | Описание | Отвергнут потому что |
|---------|----------|---------------------|
| localStorage | Сохранять URL в localStorage | Лишняя сложность, может протухнуть |
| sessionStorage | Сохранять URL в sessionStorage | Не работает между вкладками |
| Query-параметр `redirect` | Передать URL в query string | **ВЫБРАН** — простой, stateless, стандартный паттерн |

---

## Исследование 4: Увеличение TTL access token

### Decision: Изменить `access_token_ttl` с 15m на 2h во всех конфигах

### Rationale

Пользователь запросил увеличение времени жизни сессии до "пары часов". Access token TTL определяет, как часто требуется refresh. При 2h TTL пользователь может работать без единого refresh в течение 2 часов.

**Безопасность**: Для домашней библиотеки с единицами пользователей в локальной сети (§6.II конституции) 2-часовой access token — приемлемый компромисс между удобством и безопасностью. Refresh token TTL (30 дней) остаётся без изменений.

**Файлы для изменения**:
- `config/config.dev.yaml` — `access_token_ttl: "2h"`
- `config/config.stage.yaml` — `access_token_ttl: "2h"`
- `config/config.prod.yaml` — `access_token_ttl: "2h"`
- `backend/internal/config/config.go` — дефолт `2 * time.Hour`

---

## Исследование 5: Баг кнопки "Выход"

### Decision: Обработать ошибку logout gracefully

### Rationale

**Проблема**: В `AppHeader.vue`:

```typescript
async function handleLogout() {
  await auth.logout()     // ← может упасть с ошибкой, если API недоступен
  router.push('/login')   // ← не выполнится если logout throws
}
```

`auth.logout()` вызывает `authApi.logout()` → `api.post('/auth/logout')`. Если access token истёк, interceptor пытается refresh, который тоже может упасть. При этом `store.logout()` имеет `finally { clearAuth() }`, но `handleLogout` может выбросить ошибку до `router.push('/login')`.

Кроме того, AppHeader показывает меню по условию `auth.isAuthenticated`. Если store очищается (clearAuth), но навигация на login не происходит, пользователь видит пустой header без меню на странице каталога.

**Решение**: Обернуть `handleLogout` в try/catch, гарантировать очистку и редирект в любом случае.

---

## Сводная таблица дефектов

| # | Файл | Описание | Приоритет |
|---|------|----------|-----------|
| 1 | `config/config.stage.yaml:19` | `cookie_secure: true` при HTTP → cookie отклоняется | **P0 — корневая причина** |
| 2 | `frontend/src/api/client.ts:78-81` | Store не очищается при неудачном refresh → guard не пускает на login | **P1 — блокирует UX** |
| 3 | `frontend/src/views/LoginView.vue:117` | Нет redirect-back после повторного входа | P2 |
| 4 | `frontend/src/components/AppHeader.vue:45-48` | handleLogout не обрабатывает ошибки → меню пропадает | P2 |
| 5 | `config/*.yaml`, `config.go:79` | Access token TTL 15m → нужно 2h | P3 |
