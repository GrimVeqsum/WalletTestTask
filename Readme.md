# Wallet Service

REST-сервис для создания кошельков, изменения баланса и получения текущего баланса.

Стек: Go, PostgreSQL, Docker.

## Запуск

```bash
docker compose up --build -d
```

Проверка контейнеров:

```bash
docker compose ps
```

Просмотр логов приложения:

```bash
docker compose logs app
```

Остановка приложения:

```bash
docker compose down
```

Остановка с удалением данных PostgreSQL:

```bash
docker compose down -v
```

## API

### Создание кошелька

```http
POST /api/v1/wallets
```

Пример запроса:

```bash
curl -X POST http://localhost:8080/api/v1/wallets
```

Пример ответа:

```json
{
  "walletId": "dc040d51-4803-4aa3-a3ed-f6e2c52b5884",
  "balance": 0
}
```

Кошелёк создаётся отдельной ручкой, чтобы операции изменения баланса выполнялись только для существующих кошельков.

### Пополнение кошелька

```http
POST /api/v1/wallet
```

Пример запроса:

```bash
curl -X POST http://localhost:8080/api/v1/wallet \
  -H "Content-Type: application/json" \
  -d '{
    "walletId": "dc040d51-4803-4aa3-a3ed-f6e2c52b5884",
    "operationType": "DEPOSIT",
    "amount": 1000
  }'
```

Пример ответа:

```json
{
  "walletId": "dc040d51-4803-4aa3-a3ed-f6e2c52b5884",
  "balance": 1000
}
```

### Снятие средств

```http
POST /api/v1/wallet
```

Пример запроса:

```bash
curl -X POST http://localhost:8080/api/v1/wallet \
  -H "Content-Type: application/json" \
  -d '{
    "walletId": "dc040d51-4803-4aa3-a3ed-f6e2c52b5884",
    "operationType": "WITHDRAW",
    "amount": 400
  }'
```

Пример ответа:

```json
{
  "walletId": "dc040d51-4803-4aa3-a3ed-f6e2c52b5884",
  "balance": 600
}
```

Если средств недостаточно, сервис возвращает:

```text
409 Conflict
```

### Получение баланса

```http
GET /api/v1/wallets/{walletId}
```

Пример запроса:

```bash
curl http://localhost:8080/api/v1/wallets/dc040d51-4803-4aa3-a3ed-f6e2c52b5884
```

Пример ответа:

```json
{
  "walletId": "dc040d51-4803-4aa3-a3ed-f6e2c52b5884",
  "balance": 600
}
```

## Коды ответов

```text
200 OK — операция выполнена
201 Created — кошелёк создан
400 Bad Request — неправильный JSON, сумма или тип операции
404 Not Found — кошелёк не найден
409 Conflict — недостаточно средств
500 Internal Server Error — внутренняя ошибка
```

## Тесты

Полный запуск тестов с отдельной PostgreSQL в Docker:

```bash
docker compose -f docker-compose.test.yml up --build --exit-code-from test
```

Удаление тестового окружения:

```bash
docker compose -f docker-compose.test.yml down -v --remove-orphans
```

Интеграционные тесты проверяют конкурентные пополнения и снятия средств по одному кошельку. Изменение баланса выполняется атомарными SQL-запросами, что предотвращает потерю обновлений при одновременных запросах.
