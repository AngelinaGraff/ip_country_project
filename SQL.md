
### Структура таблиц MySQL

#### Таблица ip_networks:

    id (INT, Primary Key, Auto Increment)
    ip_version (TINYINT) — версия IP (4 или 6)
    network (VARBINARY(16)) — сеть в бинарном виде
    prefix_length (INT) — длина префикса сети
    country_code (VARCHAR(2)) — код страны

#### Причины выбора структуры:

##### VARBINARY(16):
- Позволяет хранить как **IPv4** (4 байта), так и **IPv6** (16 байт) адреса.
- Бинарный формат ускоряет сравнения и поиск.

##### Индексы:
- Возможно создание индексов по колонкам **ip_version, network, prefix_length** для ускорения запросов.

### SQL-запросы для поиска сети

#### Запрос для IPv4:
```sql
SELECT country_code, prefix_length
FROM ip_networks
WHERE ip_version = 4
  AND network = SUBSTRING(INET6_ATON('192.0.2.1'), 1,   prefix_length / 8)
ORDER BY prefix_length DESC
LIMIT 1;
```

#### Запрос для IPv6:

```sql
SELECT country_code, prefix_length
FROM ip_networks
WHERE ip_version = 6
  AND network = SUBSTRING(INET6_ATON('2001:db8::1'), 1, prefix_length / 8)
ORDER BY prefix_length DESC
LIMIT 1;
```

#### Объяснение:

- **INET6_ATON() :** Преобразует IP-адрес в бинарный формат.

- **SUBSTRING() :** Извлекает часть адреса, соответствующую длине префикса.

- **ORDER BY prefix_length DESC :** Сортирует по длине префикса, чтобы найти наиболее специфичную сеть.

- **LIMIT 1 :** Ограничивает результат одним значением для ускорения запроса.