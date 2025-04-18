# Avito PVZ Backend Service

Тестовое задание для стажёра Backend-направления (весенняя волна 2025)

## Содержание

- [Описание](#описание)
- [Технологии](#технологии)
- [Структура проекта](#структура-проекта)
- [Структура БД](#структура-бд)
- [Запуск](#запуск)
- [Тестирование](#тестирование)
- [Нагрузочное тестирование](#нагрузочное-тестирование)
- [HTTP-хэндлеры](#http-хэндлеры)
- [Метрики Prometheus](#метрики-prometheus)
- [gRPC-сервис](#grpc-сервис)

---

## Описание

Это бэкенд-сервис для управления работой пунктов выдачи заказов (ПВЗ). Сервис реализует:

- Регистрацию и авторизацию пользователей
- Создание и управление ПВЗ
- Приёмки и закрытие поступающих товаров
- Удаление и добавление товаров по LIFO
- Получение данных о ПВЗ с фильтром и пагинацией

## Технологии

- Go + Gin
- PostgreSQL
- Docker / Docker Compose
- JWT Authentication
- Unit tests + Integration tests
- k6 для нагрузочного тестирования

## Структура проекта

```
.
├── cmd/server/main.go              # точка входа
├── internal
│   ├── handler                    # HTTP-хэндлеры
│   ├── middleware                 # JWT проверка
│   ├── repository                 # логика работы с БД
│   └── database                   # подключение к БД
├── migrations/                    # SQL-схема
├── tests/
│   ├── integration/              # интеграционные тесты
│   └── stress/                   # нагрузочные тесты (k6)
├── .env                           # переменные среды
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## Структура БД

```
users
- id UUID PRIMARY KEY
- email VARCHAR(255) NOT NULL UNIQUE
- password VARCHAR(255) NOT NULL (хэшированный)
- role VARCHAR(50) CHECK (role IN ('client', 'moderator', 'staff'))
- created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()

pvz
- id UUID PRIMARY KEY
- registration_date TIMESTAMP WITH TIME ZONE DEFAULT NOW()
- city VARCHAR(255)

receptions
- id UUID PRIMARY KEY
- date_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
- pvz_id UUID REFERENCES pvz(id) ON DELETE CASCADE
- status VARCHAR(50) CHECK (status IN ('in_progress', 'close'))

products
- id UUID PRIMARY KEY
- date_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
- type VARCHAR(50) CHECK (type IN ('электроника', 'одежда', 'обувь'))
- reception_id UUID REFERENCES receptions(id) ON DELETE CASCADE
```

## Запуск

```bash
git clone https://github.com/fatkheev/avito-pvz-backend
cd avito-pvz-backend
cp .env.example .env
make run
```

В процессе запуска будет автоматически выполнена миграция базы данных. В отдельном контейнере Docker будет поднят PostgreSQL, в котором создадутся необходимые таблицы.

После запуска можно выполнить интеграционный тест:

```bash
make integration-test
```

Этот тест:

- создаёт новый ПВЗ
- инициализирует приёмку
- добавляет 50 товаров
- закрывает приёмку

❗ Интеграционные тесты работают только после запуска проекта через `make run`

## Тестирование

### Unit

```bash
make test          # запуск всех тестов
make cover         # генерация coverage.html
```

Стек:

- `testify`, `httptest` для unit
- `sqlmock` для мока базы

---

## Нагрузочное тестирование

### Цель

Проверить RPS, среднее время отклика и долю успешных запросов

### Требования

- RPS >= 1000
- SLI отклика < 100ms
- SLI успеха ответов — 99.99%

### Установка k6

#### Linux

```bash
sudo apt install -y gnupg software-properties-common
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 33C235A34C46AA3FFB293709A328C3A2C3C45C06
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt update && sudo apt install k6
```

#### macOS (через Homebrew)

```bash
brew install k6
```

#### Windows

Скачайте `.exe` файл с [k6.io/docs/getting-started/installation](https://k6.io/docs/getting-started/installation/) и добавьте путь в `PATH`.

### Запуск

```bash
k6 run tests/stress/stress_test.js
```

### k6 config (пример)

```javascript
export const options = {
    vus: 30,
    duration: '10s',
    thresholds: {
        http_req_failed: ['rate<0.0001'],
        http_req_duration: ['p(99)<100'],
    },
};
```

### Результат

- Запрос: `GET /pvz?startDate=2020-01-01T00:00:00Z&endDate=2030-01-01T00:00:00Z&page=1&limit=10`
- Выполнено: ~31.5k запросов за 10 секунд
- Успешных: 99.9%
- Среднее время: ~14.5ms
- P99: ~60ms

### Вывод

✔ Сервис соответствует требованиям по производительности и стабильности при нагрузке в 30 VU / 3k RPS.

---

## HTTP-хэндлеры

_(все запросы выполняются на http://localhost:8080)_

### 1. `POST /dummyLogin` **(публичный)**

Сгенерировать тестовый JWT-токен.

**Пример запроса**
```json
{
  "role": "moderator"
}
```

**Пример ответа**
```json
{
  "token": "<JWT>"
}
```

### 2. `POST /register` **(публичный)**

Регистрация нового пользователя.

**Пример запроса**
```json
{
  "email": "example@mail.ru",
  "password": "123456",
  "role": "client"
}
```

**Пример ответа**
```json
{
  "id": "...",
  "email": "example@mail.ru",
  "role": "client",
  "created_at": "..."
}
```

### 3. `POST /login` **(публичный)**

Авторизация по email и паролю.

**Пример запроса**
```json
{
  "email": "example@mail.ru",
  "password": "123456"
}
```

**Пример ответа**
```json
{
  "token": "<JWT>"
}
```

### 4. `POST /pvz` **(защищённый, только moderator)**

Создание нового ПВЗ.

**Заголовки:**
```
Authorization: Bearer <token>
```

**Пример запроса**
```json
{
  "city": "Казань"
}
```

**Пример ответа**
```json
{
  "id": "...",
  "registration_date": "...",
  "city": "Казань"
}
```

### 5. `POST /receptions` **(защищённый, только staff)**

Открытие новой приёмки.

**Заголовки:**
```
Authorization: Bearer <token>
```

**Пример запроса**
```json
{
  "pvzId": "<pvz_id>"
}
```

**Пример ответа**
```json
{
  "id": "...",
  "date_time": "...",
  "pvz_id": "...",
  "status": "in_progress"
}
```

### 6. `POST /products` **(защищённый, только staff)**

Добавление товара в приёмку.

**Заголовки:**
```
Authorization: Bearer <token>
```

**Пример запроса**
```json
{
  "pvzId": "<pvz_id>",
  "type": "электроника"
}
```

**Пример ответа**
```json
{
  "id": "...",
  "date_time": "...",
  "type": "электроника",
  "reception_id": "...",
  "pvz_id": "..."
}
```

### 7. `POST /pvz/{pvzId}/close_last_reception` **(защищённый, только staff)**

Закрытие активной приёмки.

**Заголовки:**
```
Authorization: Bearer <token>
```

**Пример ответа**
```json
{
  "id": "...",
  "status": "close"
}
```

### 8. `POST /pvz/{pvzId}/delete_last_product` **(защищённый, только staff)**

Удаление последнего товара (LIFO) из приёмки.

**Заголовки:**
```
Authorization: Bearer <token>
```

**Пример ответа**
```json
{
  "message": "Last product deleted"
}
```

### 9. `GET /pvz` **(защищённый, moderator или staff)**

Получение списка ПВЗ с фильтрацией и пагинацией.

**Query-параметры:**
```
startDate, endDate, page, limit
```

**Заголовки:**
```
Authorization: Bearer <token>
```

**Пример ответа**
```json
[
  {
    "pvz": {
      "id": "...",
      "registration_date": "...",
      "city": "Казань"
    },
    "receptions": [
      {
        "reception": {
          "id": "...",
          "date_time": "...",
          "status": "close"
        },
        "products": [
          {
            "id": "...",
            "type": "электроника",
            ...
          }
        ]
      }
    ]
  }
]
```

## Метрики Prometheus

После запуска проекта Prometheus метрики доступны по адресу: [http://localhost:9000/metrics](http://localhost:9000/metrics)

### Технические метрики

#### `http_requests_total`
Общее число HTTP-запросов

```
http_requests_total{method="POST",path="/pvz",status="201"} 1
http_requests_total{method="POST",path="/receptions",status="201"} 1
http_requests_total{method="POST",path="/products",status="201"} 1
```

#### `http_request_duration_seconds`
Гистограмма времени обработки запросов

```
http_request_duration_seconds_bucket{method="POST",path="/pvz",le="0.1"} 1
http_request_duration_seconds_sum{method="POST",path="/pvz"} 0.00266
http_request_duration_seconds_count{method="POST",path="/pvz"} 1
```

### Бизнес-метрики

#### `pvz_created_total`
Сколько раз был вызван `POST /pvz`
```
pvz_created_total 1
```

#### `receptions_created_total`
Сколько раз был вызван `POST /receptions`
```
receptions_created_total 1
```

#### `products_created_total`
Сколько раз был вызван `POST /products`
```
products_created_total 1
```

### Прочее

Все метрики `go_*`, `process_*`, `promhttp_*` — системные и относятся к мониторингу самого сервиса (потоки, память и т.д.).

---

## gRPC-сервис

В рамках задания реализован отдельный gRPC-сервер, доступный на порту `3000`. Он не требует авторизации и предназначен для получения списка всех ПВЗ в системе.

- Протокол: gRPC
- Порт: `3000`
- Сервис: `pvz.v1.PVZService`
- Метод: `GetPVZList`

### Вызов через grpcurl

```bash
grpcurl -plaintext localhost:3000 pvz.v1.PVZService.GetPVZList
```

### Ответ (пример)

```json
{
  "pvz": [
    {
      "id": "b7c47d9c-5e91-4c3b-b9d0-6aa470d0f39c",
      "city": "Москва",
      "registrationDate": "2025-04-17T07:23:10Z"
    }
  ]
}
```

gRPC реализован с использованием `protoc`, файл схемы находится в `proto/pvz.proto`.

