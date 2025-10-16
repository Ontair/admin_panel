# 🚀 Руководство по развертыванию Admin Panel

## Варианты запуска

### 1. Простой запуск (только backend)

```bash
cd /Users/ontair/GoLang/admin_panel
docker-compose up -d
```

**Доступ:**
- Backend API: http://localhost:8080
- Health check: http://localhost:8080/health

### 2. Запуск с Nginx (рекомендуется для production)

```bash
cd /Users/ontair/GoLang/admin_panel
docker-compose -f docker-compose.nginx.yml up -d
```

**Доступ:**
- Backend API через Nginx: http://localhost
- Health check: http://localhost/health
- Login API: http://localhost/api/v1/auth/login

**Преимущества:**
- Nginx обрабатывает CORS headers
- Лучшая производительность
- Готовность к production

### 3. Запуск с Traefik (для production с SSL)

```bash
cd /Users/ontair/GoLang/admin_panel
docker-compose -f docker-compose.traefik.yml up -d
```

**Доступ:**
- Backend API через Traefik: http://localhost
- Traefik Dashboard: http://localhost:8080
- Health check: http://localhost/health

**Преимущества:**
- Автоматическое SSL (Let's Encrypt)
- Service discovery
- Load balancing
- Мониторинг через dashboard

## 🔧 Настройка для production

### Переменные окружения

Для production измените переменные в docker-compose файлах:

```yaml
environment:
  - DATABASE_HOST=your-db-host
  - DATABASE_PORT=5432
  - DATABASE_USERNAME=your-db-user
  - DATABASE_PASSWORD=your-secure-password
  - DATABASE_NAME=admin_panel
```

### CORS настройки

Для production измените CORS в `cmd/main.go`:

```go
c.Header("Access-Control-Allow-Origin", "https://your-frontend-domain.com")
```

И в `nginx.conf` или Traefik labels соответственно.

### SSL сертификаты

#### С Nginx:
1. Получите SSL сертификат (Let's Encrypt)
2. Обновите `nginx.conf` для HTTPS
3. Настройте редирект с HTTP на HTTPS

#### С Traefik:
1. Добавьте Let's Encrypt resolver в `traefik.yml`
2. Добавьте labels для SSL в `docker-compose.traefik.yml`

## 🧪 Тестирование

### Проверка health check
```bash
curl http://localhost/health
```

### Тестирование login
```bash
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'
```

### Проверка CORS
```bash
curl -H "Origin: http://localhost:5173" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS http://localhost/api/v1/auth/login
```

## 📊 Мониторинг

### Логи
```bash
# Все сервисы
docker-compose logs -f

# Только backend
docker-compose logs -f backend

# Только Nginx
docker-compose -f docker-compose.nginx.yml logs -f nginx

# Только Traefik
docker-compose -f docker-compose.traefik.yml logs -f traefik
```

### Статус контейнеров
```bash
docker-compose ps
```

### Traefik Dashboard
- URL: http://localhost:8080
- Показывает все роуты, сервисы и middlewares

## 🔄 Обновление

### Пересборка и перезапуск
```bash
# Простой запуск
docker-compose up --build -d

# С Nginx
docker-compose -f docker-compose.nginx.yml up --build -d

# С Traefik
docker-compose -f docker-compose.traefik.yml up --build -d
```

### Остановка
```bash
# Остановить все
docker-compose down

# Остановить с Nginx
docker-compose -f docker-compose.nginx.yml down

# Остановить с Traefik
docker-compose -f docker-compose.traefik.yml down
```

## 🐛 Troubleshooting

### Backend не запускается
1. Проверьте логи: `docker-compose logs backend`
2. Убедитесь, что PostgreSQL запущен: `docker-compose logs postgres`
3. Проверьте переменные окружения

### CORS ошибки
1. Убедитесь, что CORS настроен для вашего frontend домена
2. Проверьте, что `Access-Control-Allow-Credentials: true`
3. Для Nginx: проверьте `nginx.conf`
4. Для Traefik: проверьте labels в docker-compose

### База данных
1. Проверьте подключение: `docker-compose exec postgres psql -U postgres -d admin_panel`
2. Проверьте таблицы: `\dt`
3. Пересоздайте БД: `docker-compose down -v && docker-compose up -d`

## 📝 Дополнительные команды

### Создание пользователя через API
```bash
curl -X POST http://localhost/api/v1/manager/users \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=your-token" \
  -d '{
    "username": "newuser",
    "password": "password123",
    "first_name": "New",
    "last_name": "User",
    "role": "user",
    "is_active": true
  }'
```

### Получение списка пользователей
```bash
curl -H "Cookie: access_token=your-token" \
     http://localhost/api/v1/manager/users
```

## 🔐 Безопасность

### Рекомендации для production:
1. Измените JWT секреты в `config.yaml`
2. Используйте сильные пароли для БД
3. Настройте HTTPS
4. Ограничьте доступ к Traefik dashboard
5. Используйте secrets для паролей в Docker
6. Настройте firewall правила
7. Регулярно обновляйте зависимости



