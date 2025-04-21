# Сервис для работы с ПВЗ

## Описание проекта

Сервис для работы с пунктами выдачи заказов (ПВЗ), который позволяет:
- Создавать ПВЗ (только модераторам)
- Создавать приемки товаров
- Добавлять товары в рамках приемки
- Удалять товары (по принципу LIFO)
- Закрывать приемки
- Получать информацию о ПВЗ и товарах

## Конфигурация

Приложение использует `.env` файл и переменные окружения для конфигурации. 

### Настройка .env файла

1. Скопируйте файл примера конфигурации:
```bash
cp .env.example .env
```

2. Отредактируйте `.env` файл, настроив нужные параметры:
```
# Сервер
HTTP_ADDR=:8080                # Порт HTTP сервера
GRPC_ADDR=:3000                # Порт gRPC сервера
PROMETHEUS_ADDR=:9000          # Порт для метрик Prometheus
SHUTDOWN_TIMEOUT=5s            # Таймаут для graceful shutdown

# База данных
# URL подключения к БД (для локального запуска используйте localhost вместо db)
DB_URL=postgres://postgres:postgres@db:5432/avito?sslmode=disable
DB_MAX_CONN=10                 # Максимальное количество соединений
DB_MIN_CONN=5                  # Минимальное количество соединений

# PostgreSQL
POSTGRES_USER=postgres         # Имя пользователя PostgreSQL
POSTGRES_PASSWORD=postgres     # Пароль PostgreSQL
POSTGRES_DB=avito              # Имя базы данных

# JWT
JWT_SECRET=supersecretkey      # Секретный ключ для подписи JWT
TOKEN_TTL=24h                  # Время жизни токена

# Логирование
LOG_LEVEL=info                 # Уровень логирования (debug, info, warn, error)
```

## Запуск проекта

### С использованием Docker Compose

```bash
# Настройка переменных окружения
cp .env.example .env
# При необходимости отредактируйте .env файл

# Запуск сервиса и базы данных
docker-compose up -d
```

После запуска сервис будет доступен:
- HTTP API: http://localhost:8080
- gRPC API: localhost:3000
- Prometheus метрики: http://localhost:9000/metrics


## API Endpoints

### Аутентификация
- `POST /dummyLogin` - Получение тестового токена
- `POST /register` - Регистрация нового пользователя
- `POST /login` - Авторизация по email и паролю

### ПВЗ
- `POST /pvz` - Создание ПВЗ (только для модераторов)
- `GET /pvz` - Получение списка ПВЗ
- `GET /pvz/{id}` - Получение информации о ПВЗ по ID

### Приемки
- `POST /receptions` - Создание новой приемки
- `POST /pvz/{pvzId}/close_last_reception` - Закрытие последней приемки

### Товары
- `POST /products` - Добавление товара в текущую приемку
- `POST /pvz/{pvzId}/delete_last_product` - Удаление последнего добавленного товара

## Дополнительные возможности

1. gRPC сервис - доступен на порту 3000:
   - `GetPVZList` - получение списка ПВЗ
   

2. Prometheus метрики - доступны на http://localhost:9000/metrics:
   - Технические метрики: количество запросов, время ответа
   - Бизнесовые метрики: количество созданных ПВЗ, приемок, товаров 

3. Кодогенерация DTO из OpenAPI-спецификации:
   ```bash
   # Генерация типов данных из OpenAPI-спецификации
   make gen-api
   ```
   
   Спецификация API находится в файле `api/openapi/v1/swagger.yaml`. 
   При запуске команды `make gen-api` генерируются Go-типы в файле `internal/interfaces/http/dto/models.gen.go`,
   которые автоматически используются обработчиками API. 