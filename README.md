# vk-normalizer

## Что делает

1. Инициализирует zap-логгер, S3-аплоадер, Kafka producer/consumer и VK-парсер по переменным окружения.
2. Подписывается на топик `TOPIC_NAME_VK_UPDATES`, десериализует `internal/vk.Update`, извлекает текст и вложения, формирует `contract.NormalizedMessage`.
3. Для каждого вложения типа `photo` скачивает файл по URL, выбирает самую крупную картинку, загружает в S3 (`internal/s3.Uploader`) и сохраняет `S3URL`.
4. Сериализует обновлённое сообщение в JSON и отправляет в топик `TOPIC_NAME_NORMALIZED_MSG` через Kafka (`internal/messaging.Producer`).

## Запуск

1. Настройте переменные окружения (ниже).
2. Соберите и запустите локально:
   ```bash
   go run ./cmd/vk-normalizer
   ```
3. Или соберите Docker-образ и запустите с нужными переменными, пример:
   ```bash
   docker build -t vk-normalizer .
   docker run --rm -e ... vk-normalizer
   ```

## Переменные окружения

Каждое значение ожидается, кроме SASL-параметров, если Kafka доступна открыто.

- `KAFKA_BOOTSTRAP_SERVERS_VALUE` — `host:port` для брокера.
- `KAFKA_GROUP_ID_VK_NORMALIZER` — идентификатор consumer group (используется при подписке на входной топик).
- `KAFKA_TOPIC_NAME_VK_UPDATES` — топик входных обновлений ВКонтакте.
- `KAFKA_TOPIC_NAME_NORMALIZED_MSG` — топик, куда публикуется нормализованное сообщение (значение обязано быть непустым).
- `KAFKA.Client_ID_VK_NORMALIZER` — общий client id для producer/consumer, отражается в метриках Sarama.
- `KAFKA_SASL_USERNAME` и `KAFKA_SASL_PASSWORD` — имя/пароль для SCRAM-SHA512 (оставьте пустыми, если не требуется).
- `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET` — параметры MinIO/S3.
- `S3_USE_SSL` — включает HTTPS (по умолчанию `false`).

## Примечания

- Парсер `internal/vk/parser.go` сортирует доступные размеры фотографий по площади и берёт самый крупный URL для загрузки.
- `internal/s3.Uploader` создаёт уникальный ключ через UUID и сохраняет оригинальное имя файла в metadata.
- Все сообщения логируются через zap, Kafka producer работает с `Snappy`, SCRAM SHA512 и идемпотентностью.
- `contract.NormalizedMessage` хранит исходный update (как `json.RawMessage`), текст, список медиа и метаинформацию (`internal/contract/normalized.go`).
- Обработка потребляет сообщения по одному; ошибки логируются, Kafka не смещает оффсет, если handler возвращает ошибку.
