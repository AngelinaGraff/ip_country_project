# Проект: Сервисы для определения страны по IP-адресу

## Описание проекта

Данный проект включает в себя два сервиса, реализованных на языках программирования **Go** и **PHP**, которые предоставляют API для определения страны по заданному IP-адресу (IPv4 или IPv6). Сервисы используют базу данных **GeoIP2** для получения информации о стране и реализуют кэширование результатов с помощью **Redis** для повышения производительности.

## Функциональность

1. **Определение страны по IP-адресу:**
   - Принимает IP-адрес через GET-параметр.
   - Использует библиотеку GeoIP2 для получения информации о стране.
   - Кэширует результаты в Redis для ускорения повторных запросов.

## Технологии и инструменты

- **Go**: Выбран для создания высокопроизводительного сервиса.
- **PHP**: Широко используемый язык для веб-разработки, обеспечивает быстрый старт и простоту поддержки.
- **GeoIP2**: Предоставляет актуальные и точные данные о геолокации IP-адресов.
- **Redis**: Используется для кэширования результатов запросов, что значительно снижает нагрузку на базу данных GeoIP2 и повышает скорость отклика сервисов.
- **Docker** и **Docker Compose**: Обеспечивают контейнеризацию приложений для легкого развертывания и масштабирования.
- **Composer** и **Go Modules**: Используются для управления зависимостями в PHP и Go проектах соответственно.

## Структура проекта

- `go_service/`: Сервис на Go.
- `php_service/`: Сервис на PHP.

## Установка и запуск

### Предварительные требования

- Docker
- Docker Compose

### Шаги по запуску

1. **Клонируйте репозиторий:**

    ```bash
    git clone https://github.com/AngelinaGraff/ip_country_project.git
    ```
2. **Перейдите в директорию проекта:**
    ```bash
    cd ip_country_project
    ```
3. **Запустите Docker Compose для Go сервиса:**
    ```bash
    cd .\go_service\
    docker-compose up -d
    ```
4. **Запустите Docker Compose для PHP сервиса:**
    ```bash
    docker build -t geoip-service-php
    docker-compose up -d
    ```

## Использование

### Go Сервис

- **Эндпоинт для получения страны по IP:**

    ```bash
    GET http://localhost:8080/getcountry?ip=<IP_ADDRESS>
    ```

- **Пример запроса:**

    ```bash
    http://localhost:8080/getcountry?ip=8.8.8.8
    ```
- **Пример ответа:**

    ```json
    {
      "country": "US"
    }
    ```
### PHP Сервис

- Эндпоинт для получения страны по IP:

    ```bash
    GET http://localhost:8081/index.php?ip=<IP_ADDRESS>
    ```

- Пример запроса:
    
    ```bash
    http://localhost:8081/index.php?ip=8.8.8.8
    ```
- Пример ответа:

    ```json
    {
      "ip": "8.8.8.8",
      "country": {
        "iso_code": "US",
        "name": "United States"
      }
    }
    ```

## Технические детали
### Выбор технологий

### Go:
#### Причины выбора:
1. Высокая производительность и эффективность.
2. Компилируется в один бинарный файл, что упрощает развертывание.

#### Использование:
1. Реализован HTTP-сервис для обработки запросов.
2. Используется пакет geoip2-golang для работы с базой GeoIP2.
3. Кэширование реализовано с помощью клиента Redis go-redis/redis/v8.

### PHP:
#### Причины выбора:
1. Широко распространен и прост в использовании.
2. Большое количество библиотек и сообществ.
#### Использование:
1. Реализован веб-сервис на основе Apache.
3. Используется библиотека geoip2/geoip2 для работы с GeoIP2.
1. Кэширование реализовано с помощью predis/predis.

### GeoIP2:
#### Причины выбора:
1. Предоставляет актуальные и точные данные о геолокации IP-адресов.
2. Поддерживает как IPv4, так и IPv6.
#### Использование:
1. База данных GeoLite2-Country.mmdb используется для определения страны по IP.

### Redis:
#### Причины выбора:
1. Быстрое хранение данных в памяти.
2. Поддерживает различные структуры данных.
3. Идеально подходит для кэширования часто запрашиваемых данных.
#### Использование:
1. Кэширование результатов запросов к GeoIP2 для снижения нагрузки и повышения скорости.

### Кэширование

#### Причины использования кэширования:
1. Снижение нагрузки на базу данных **GeoIP2**.
2. Ускорение времени отклика сервисов.

#### Стратегия кэширования:
1. Результаты запросов кэшируются в Redis с TTL (Time To Live).
2. Для Go-сервиса TTL установлено на 24 часа, для PHP-сервиса — на 1 час.
3. Причины различий в TTL:
    - В Go-сервисе предполагается более редкое обновление данных.
    - В PHP-сервисе возможны более частые обновления, поэтому TTL меньше.

### Конфигурация

#### Внешние файлы конфигурации:
- Для Go — config.yaml.
- Для PHP — config.php.
#### Причины использования:
- Облегчает изменение настроек без необходимости изменения кода.
- Повышает гибкость и адаптивность приложения.

### Docker и контейнеризация

#### Причины использования Docker:
- Обеспечивает изоляцию сервисов.
- Упрощает развертывание и масштабирование.
- Гарантирует согласованность окружения в разных средах.

#### Docker Compose:
- Управляет множественными контейнерами как единым приложением.
- Облегчает настройку сети между сервисами и Redis.

### Развертывание

1. Сервисы готовы к запуску в любой среде, поддерживающей Docker.
2. Конфигурационные файлы позволяют легко адаптировать приложения под разные окружения.

#### Тестирование

docker-compose run --rm test