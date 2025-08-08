## подсказка

В производственных средах мы настоятельно рекомендуем запускать ClickHouse Keeper на выделенных узлах. В тестовых средах, если вы решите запускать ClickHouse
Server и ClickHouse Keeper на одном сервере, то вам не нужно устанавливать ClickHouse Keeper, так как он уже включен в сервер ClickHouse.

https://clickhouse.com/docs/ru/install/debian_ubuntu#install-standalone-clickhouse-keeper

## ClickHouse Cluster Deployment
Этот репозиторий автоматизирует развёртывание отказоустойчивого кластера ClickHouse с использованием ClickHouse Keeper для координации. 
Кластер поддерживает шардирование данных и репликацию.

# Структура директории

```text
simple-clickhouse-cluster/
├── cluster.yaml              # Основной плейбук развертывания
├── group_vars/
│   └── all.yml               # Глобальные переменные
├── inventory.yaml            # Пример инвентарного файла
├── roles/
│   ├── install_CH/           # Установка ClickHouse Server
│   │   ├── handlers/
│   │   │   └── main.yml      # Обработчики сервиса ClickHouse
│   │   └── tasks/
│   │       ├── configure.yaml # Конфигурация ClickHouse
│   │       ├── install.yaml   # Установка пакетов
│   │       └── main.yaml      # Основные задачи
│   └── install_keeper/       # Установка ClickHouse Keeper
│       ├── handlers/
│       │   └── main.yml      # Обработчики сервиса Keeper
│       └── tasks/
│           └── main.yaml     # Установка и настройка Keeper
└── templates/
    ├── ch/                   # Шаблоны ClickHouse
    │   ├── client/           # Конфиги клиента
    │   │   └── config.xml
    │   ├── conf/             # Основные конфиги сервера
    │   │   ├── cluster.xml   # Конфиг кластера
    │   │   ├── data_path.xml # Пути данных
    │   │   ├── data_system.xml # Системные таблицы
    │   │   ├── logger.xml    # Логирование
    │   │   ├── macros/       # Макросы для репликации
    │   │   │   └── macros.xml
    │   │   ├── ports.xml     # Портовая конфигурация
    │   │   ├── prometheus.xml # Мониторинг
    │   │   ├── user_directories.xml # Аутентификация
    │   │   └── zoo.xml       # Конфиг ZooKeeper/Keeper
    │   ├── config.xml        # Главный конфиг
    │   └── users.xml         # Пользователи и права
    └── keeper/               # Шаблоны Keeper
        └── config.xml        # Конфиг ClickHouse Keeper
```

## Ключевые особенности

  1. Автоматическое шардирование и репликация
  2. Количество шардов определяется по формуле clickhouse_servers/clickhouse_replicas=shards
  3. Конфиг кластера генерируется динамически на основе инвентаря:
```jinja
<remote_servers>
    <{{ clickhouse_cluster }}>
    {% for batch in groups[clickhouse_servers] | batch(clickhouse_replicas) %}
    <shard_{{ loop.index }}>
        <weight>{{ clickhouse_replicas }}</weight>
        <internal_replication>true</internal_replication>
        {% for host in batch %}
        <replica>
            <host>{{ host }}</host>
            <port>{{ clickhouse_native_port}}</port>
        </replica>
        {% endfor %}
    </shard_{{ loop.index }}>
    {% endfor %}
</{{ clickhouse_cluster }}>
```

## 2. Автоматическое назначение макросов

```jinja
<clickhouse>
   <macros>
{% for batch in groups['clickhouse_servers'] | batch(clickhouseReplicas, fill_with=None) %}
{% set shard_id = loop.index  %}
{% for host in batch %}
{% if host == inventory_hostname %}
{% set short_hostname = inventory_hostname.split('.')[0] %}
      <shard>shard_{{ shard_id }}</shard>
      <replica>replica_{{ ansible_host | ipaddr('last_octet') }}</replica>
      <cluster>{{ logs_cluster }}</cluster>
{% endif %}
{% endfor %}
{% endfor %}
   </macros>
</clickhouse>
```
## 3 4. Мониторинг и логирование
   Встроенная поддержка Prometheus

   Расширенное логирование в JSON-формате

   Настройка TTL для системных таблиц

# Переменные (group_vars/all.yml)

Основные параметры:

```yaml
# Версия ClickHouse
clickhouse_repo: "deb https://packages.clickhouse.com/deb stable main"

# Параметры кластера
# Количество шардов определяется по формуле clickhouse_servers/clickhouse_replicas=shards
clickhouse_replicas: 2       # Реплик на шард
clickhouse_cluster: cluster  # Имя кластера

# Портовая конфигурация
clickhouse_prometheus_port: 9363
clickhouse_http_port: 8123
clickhouse_native_port: 9000
clickhouse_interserver_http_port: 9010

# Параметры Keeper
keeper_port: 9181
keeper_operation_timeout_ms: 10000
keeper_session_timeout_ms: 30000
```

## Использование

1. Подготовка инвентаря
   
Пример inventory.yaml:

```yaml
all:
  vars:
    clickhouseReplicas: 2
    clickhouse_distributed_secret: "SecurePassword123"
  children:
    clickhouse_servers:
      hosts:
        ch-node1: ansible_host=10.0.0.101
        ch-node2: ansible_host=10.0.0.102
        ch-node3: ansible_host=10.0.0.103
        ch-node4: ansible_host=10.0.0.104
    keeper_servers:
      hosts:
        keeper1: ansible_host=10.0.0.201
        keeper2: ansible_host=10.0.0.202
        keeper3: ansible_host=10.0.0.203
```

## 2. Запуск развёртывания

```bash
ansible-playbook -i inventory.yaml cluster.yaml
```

# Этапы выполнения:
   1. Установка ClickHouse Keeper на ноды из группы keeper_servers
   2. Установка ClickHouse Server на ноды из группы clickhouse_servers
   3. Настройка координации через Keeper
   4. Конфигурация шардирования и репликации
   5. Запуск и валидация кластера

## 3. Мониторинг
Встроенная поддержка Prometheus:
```jinja
<prometheus>
    <endpoint>/metrics</endpoint>
    <port>{{ clickhouse_prometheus_port }}</port>
    <metrics>true</metrics>
    <events>true</events>
</prometheus>
```


## Требования

Ansible 2.12+

Python 3.8+

Минимум 4GB RAM на ноду

SSH-доступ с правами sudo
    
    
