# Admin Panel - Hexagonal Architecture

Admin Panel - это веб-приложение для управления пользователями, построенное с использованием Hexagonal Architecture (Ports & Adapters).

## Быстрый старт

### Локальная разработка

1. **Запуск с Docker Compose (рекомендуется):**
   ```bash
   # Простой запуск
   docker-compose up -d
   
   # С Nginx (для production-like окружения)
   docker-compose -f docker-compose.nginx.yml up -d
   
   # С Traefik (для production с автоматическим SSL)
   docker-compose -f docker-compose.traefik.yml up -d
   ```

2. **Ручной запуск:**
   ```bash
   # Установка зависимостей
   go mod download
   
   # Запуск PostgreSQL (если не используется Docker)
   # Настройте config.yaml с параметрами БД
   
   # Запуск приложения
   go run cmd/main.go
   ```

### Переменные окружения

Приложение использует `config.yaml` для конфигурации. Для Docker можно переопределить через переменные окружения:

- `DB_HOST` - хост базы данных (по умолчанию: localhost)
- `DB_PORT` - порт базы данных (по умолчанию: 5432)
- `DB_USER` - пользователь БД (по умолчанию: postgres)
- `DB_PASSWORD` - пароль БД (по умолчанию: password)
- `DB_NAME` - имя БД (по умолчанию: admin_panel)

### API Endpoints

После запуска API доступно на `http://localhost:8080`:

- **Health Check:** `GET /health`
- **Auth:** `POST /api/v1/auth/login`, `POST /api/v1/auth/logout`, `POST /api/v1/auth/refresh`
- **Users:** `GET /api/v1/users/profile`, `POST /api/v1/users/change-password`
- **Manager Routes:** `GET /api/v1/manager/users`, `POST /api/v1/manager/users`
- **Admin Routes:** `GET /api/v1/admin/users`, `DELETE /api/v1/admin/users/:id`

### CORS

Backend настроен для работы с frontend на `http://localhost:5173`. Для production измените CORS настройки в `cmd/main.go`.

## Архитектура

Проект следует принципам Hexagonal Architecture:

```
cmd/
├── main.go # Точка входа приложения

internal/
├── core/
│   ├── entities/          # Доменные сущности
│   │   ├── user.go        # Пользователь с бизнес-логикой
│   │   └── errors.go      # Доменные ошибки
│   ├── dto/               # Data Transfer Objects
│   │   ├── user_dto.go    # DTO для пользователей
│   │   └── errors.go      # API ошибки
│   ├── ports/             # Интерфейсы (порты)
│   │   ├── repository/    # Репозитории
│   │   │   └── user_repository.go
│   │   └── service/       # Сервисы
│   │       ├── auth_service.go
│   │       ├── user_service.go
│   │       ├── jwt_service.go
│   │       ├── cookie_service.go
│   │       └── logger.go
│   └── services/          # Реализация бизнес-логики
│       ├── auth_service.go
│       └── user_service.go
├── infra/                 # Инфраструктура
│   ├── config/           # Конфигурация
│   ├── database/         # База данных
│   ├── jwt/              # JWT токены
│   ├── logger/           # Логирование
│   └── cookie/           # Cookie сервис
└── adapters/             # Адаптеры
    ├── primary/          # Входящие адаптеры
    │   ├── api/          # HTTP handlers
    │   │   ├── auth_handler.go
    │   │   └── user_handler.go
    │   └── middleware/   # Middleware
    │       └── auth_middleware.go
    └── secondary/        # Исходящие адаптеры (реэкспорт)
        ├── database/
        ├── jwt/
        └── cookie/
```

## Технологии

- **Go 1.21+**
- **Gin** - HTTP веб-фреймворк
- **PostgreSQL** - база данных
- **pgx** - драйвер PostgreSQL
- **JWT** - аутентификация
- **Viper** - конфигурация
- **Zap** - структурированное логирование
- **bcrypt** - хеширование паролей

## Функциональность

### Аутентификация
- Регистрация пользователей
- Вход в систему
- JWT токены (access + refresh)
- Cookie-based аутентификация
- Выход из системы

### Управление пользователями
- CRUD операции с пользователями
- Роли: Admin, Manager, User
- Активация/деактивация пользователей
- Смена пароля
- Поиск и фильтрация пользователей

### Безопасность
- Middleware для аутентификации
- Авторизация по ролям
- Валидация входных данных
- Защита от CSRF

## Установка и запуск

### Предварительные требования

1. **Go 1.21+**
2. **PostgreSQL 12+**

### Установка

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd admin_panel
```

2. Установите зависимости:
```bash
go mod tidy
```

3. Настройте базу данных:
```bash
# Создайте базу данных
createdb admin_panel

# Или используйте psql
psql -U postgres -c "CREATE DATABASE admin_panel;"
```

4. Настройте конфигурацию в `config.yaml`:
```yaml
database:
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "your_password"
  name: "admin_panel"
  sslmode: "disable"
```

### Запуск

1. **Разработка:**
```bash
go run ./cmd/main.go
```

2. **Продакшн:**
```bash
go build -o admin_panel ./cmd/main.go
./admin_panel
```

Сервер запустится на `http://localhost:8080`

## API Endpoints

### 🔓 Публичные (без аутентификации)

#### Аутентификация

**POST** `/api/v1/auth/login` - Вход в систему
```json
// Запрос
{
  "username": "admin",
  "password": "admin123"
}

// Ответ
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "username": "admin",
      "first_name": "Admin",
      "last_name": "User",
      "role": "admin",
      "is_active": true
    },
    "expires_in": 15
  },
  "message": "Login successful"
}
```
> ⚠️ Токены устанавливаются в HTTP cookies

---

**POST** `/api/v1/auth/register` - Регистрация
```json
// Запрос
{
  "username": "newuser",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}

// Ответ
{
  "success": true,
  "data": {
    "user": {
      "id": 2,
      "username": "newuser",
      "first_name": "John",
      "last_name": "Doe",
      "role": "user",
      "is_active": true
    }
  },
  "message": "User created successfully"
}
```

---

**POST** `/api/v1/auth/refresh` - Обновление токена
```json
// Ответ
{
  "success": true,
  "data": {
    "user": { ... },
    "expires_in": 15
  },
  "message": "Token refreshed successfully"
}
```
> 🔄 Refresh токен берется из cookie, новые токены устанавливаются в cookies

---

**POST** `/api/v1/auth/logout` - Выход
```json
// Ответ
{
  "success": true,
  "message": "Logged out successfully"
}
```
> 🗑️ Очищает все аутентификационные cookies

---

#### Системные

**GET** `/health` - Проверка здоровья
```json
// Ответ
{
  "status": "ok",
  "time": "2025-10-09T17:45:58.703+03:00",
  "version": "1.0.0"
}
```

---

### 🔒 Защищенные (требуют аутентификации)

#### Профиль пользователя

**GET** `/api/v1/auth/profile` - Профиль пользователя
```json
// Ответ
{
  "data": {
    "id": 1,
    "role": "admin",
    "username": "admin"
  },
  "success": true
}
```

---

**GET** `/api/v1/users/profile` - Текущий пользователь (полная информация)
```json
// Ответ
{
  "success": true,
  "data": {
    "id": 1,
    "username": "admin",
    "first_name": "Admin",
    "last_name": "User",
    "role": "admin",
    "is_active": true,
    "last_login": "2025-10-09T17:32:31.356799+03:00",
    "created_at": "2025-10-08T12:02:22.621182Z",
    "updated_at": "2025-10-09T11:23:27.226758Z"
  }
}
```

---

**POST** `/api/v1/users/change-password` - Смена пароля
```json
// Запрос
{
  "current_password": "oldpassword",
  "new_password": "newpassword123"
}

// Ответ
{
  "success": true,
  "message": "Password changed successfully"
}
```

---

### 👥 Manager+ (Manager и Admin)

#### Управление пользователями

**GET** `/api/v1/manager/users/` - Список пользователей
```
Query параметры:
- limit: 20 (по умолчанию)
- offset: 0 (по умолчанию)
- role: admin|manager|user|guest
- search: текст для поиска
- is_active: true|false
```

```json
// Ответ
{
  "users": [
    {
      "id": 1,
      "username": "admin",
      "first_name": "Admin",
      "last_name": "User",
      "role": "admin",
      "is_active": true
    }
  ],
  "total": 10,
  "limit": 20,
  "offset": 0
}
```

---

**GET** `/api/v1/manager/users/:id` - Пользователь по ID
```json
// Ответ
{
  "success": true,
  "data": {
    "id": 1,
    "username": "admin",
    "first_name": "Admin",
    "last_name": "User",
    "role": "admin",
    "is_active": true
  }
}
```

---

**PUT** `/api/v1/manager/users/:id` - Обновление пользователя
```json
// Запрос
{
  "username": "newusername",
  "first_name": "New",
  "last_name": "Name",
  "role": "manager",
  "is_active": true
}

// Ответ
{
  "success": true,
  "data": {
    "id": 1,
    "username": "newusername",
    "first_name": "New",
    "last_name": "Name",
    "role": "manager",
    "is_active": true
  }
}
```

---

### 👑 Admin только

#### Административные функции

**POST** `/api/v1/admin/users/` - Создание пользователя
```json
// Запрос
{
  "username": "newuser",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe",
  "role": "user",
  "is_active": true
}

// Ответ
{
  "success": true,
  "data": {
    "id": 3,
    "username": "newuser",
    "first_name": "John",
    "last_name": "Doe",
    "role": "user",
    "is_active": true
  }
}
```

---

**DELETE** `/api/v1/admin/users/:id` - Удаление пользователя
```json
// Ответ
{
  "success": true,
  "message": "User deleted successfully"
}
```

---

**POST** `/api/v1/admin/users/:id/activate` - Активация пользователя
```json
// Ответ
{
  "success": true,
  "message": "User activated successfully"
}
```

---

**POST** `/api/v1/admin/users/:id/deactivate` - Деактивация пользователя
```json
// Ответ
{
  "success": true,
  "message": "User deactivated successfully"
}
```

---

### 🎭 Роли и права доступа

| Роль | Описание | Доступные endpoints |
|------|----------|-------------------|
| **Guest** | Гость | Только публичные endpoints |
| **User** | Пользователь | Публичные + профиль + смена пароля |
| **Manager** | Менеджер | Все User + управление пользователями (просмотр, обновление) |
| **Admin** | Администратор | Все Manager + создание, удаление, активация/деактивация |

### 📊 Коды ответов

| Код | Описание |
|-----|----------|
| `200` | Успешно |
| `201` | Создано |
| `400` | Неверный запрос |
| `401` | Не авторизован |
| `403` | Доступ запрещен |
| `404` | Не найдено |
| `409` | Конфликт (пользователь уже существует) |
| `500` | Внутренняя ошибка сервера |

## Дефолтные учетные данные

При первом запуске создается администратор:
- **Username:** `admin`
- **Password:** `admin123`

⚠️ **Важно:** Смените пароль администратора в продакшене!

## Конфигурация

Все настройки в файле `config.yaml`:

```yaml
server:
  port: "8080"
  environment: "development"

database:
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "password"
  name: "admin_panel"

jwt:
  secret_key: "your-secret-key"
  refresh_secret: "your-refresh-secret"
  access_expiry: 15    # минуты
  refresh_expiry: 1440 # минуты (24 часа)

cookie:
  secure: false        # true для HTTPS
  same_site: "Lax"
```

## Разработка

### Структура проекта

Проект следует принципам Clean Architecture:

- **Entities** - доменные объекты с бизнес-логикой
- **Ports** - интерфейсы для внешних зависимостей
- **Services** - реализация бизнес-логики
- **Adapters** - подключение к внешним системам

### Добавление новых функций

1. Определите доменную сущность в `core/entities/`
2. Создайте порт в `core/ports/`
3. Реализуйте сервис в `core/services/`
4. Создайте адаптер в `infra/`
5. Добавьте HTTP handler в `adapters/primary/api/`

### Тестирование

```bash
# Запуск тестов
go test ./...

# Запуск с покрытием
go test -cover ./...

# Запуск конкретного теста
go test ./internal/core/services
```

## Развертывание

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o admin_panel ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/admin_panel .
COPY --from=builder /app/config.yaml .
CMD ["./admin_panel"]
```

### Переменные окружения

Можно переопределить настройки через переменные окружения:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USERNAME=postgres
export DB_PASSWORD=password
export DB_NAME=admin_panel
export JWT_SECRET_KEY=your-secret-key
```

## Безопасность

- Все пароли хешируются с помощью bcrypt
- JWT токены имеют ограниченное время жизни
- Refresh токены для безопасного обновления
- Валидация всех входных данных
- Защита от SQL инъекций через параметризованные запросы

## Мониторинг

- Структурированные логи через Zap
- Health check endpoint
- Graceful shutdown
- Метрики производительности

## Лицензия

MIT License

## Поддержка

Для вопросов и предложений создавайте Issues в репозитории.