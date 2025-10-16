# 🚀 Быстрый запуск Admin Panel

## 1. Запуск Backend

```bash
cd /Users/ontair/GoLang/admin_panel

# Выберите один из вариантов:

# Простой запуск (порт 8080)
docker-compose up -d

# Или с Nginx (порт 80)
docker-compose -f docker-compose.nginx.yml up -d

# Или с Traefik (порт 80)
docker-compose -f docker-compose.traefik.yml up -d
```

## 2. Запуск Frontend

```bash
cd /Users/ontair/front/admin_panel

# Установка зависимостей (если нужно)
npm install

# Запуск в режиме разработки
npm run dev
```

## 3. Проверка

- **Frontend:** http://localhost:5173 (или 5174, если 5173 занят)
- **Backend API:** 
  - Простой: http://localhost:8080
  - С Nginx/Traefik: http://localhost

## 4. Вход в систему

- **Username:** `admin`
- **Password:** `admin123`

## 5. Остановка

```bash
# Backend
cd /Users/ontair/GoLang/admin_panel
docker-compose down

# Frontend
# Нажмите Ctrl+C в терминале где запущен npm run dev
```

## 🔧 Настройка API URL

Если frontend запустился на другом порту, обновите `.env.development`:

```bash
cd /Users/ontair/front/admin_panel
echo "VITE_API_URL=http://localhost" > .env.development
```

## 🐛 Если не работает

1. **Проверьте порты:**
   - Backend: `curl http://localhost/health`
   - Frontend: откройте http://localhost:5173

2. **Проверьте CORS:**
   - Убедитесь, что CORS настроен для правильного порта frontend

3. **Перезапустите:**
   ```bash
   # Backend
   docker-compose down && docker-compose up -d
   
   # Frontend
   # Остановите и запустите заново npm run dev
   ```

## 📋 Что должно работать

- ✅ Вход в систему
- ✅ Dashboard
- ✅ Profile
- ✅ Users управление (для admin/manager)
- ✅ Logout



