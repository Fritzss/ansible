# Keepalived HA Cluster Role

Эта роль автоматизирует развертывание отказоустойчивого кластера с использованием Keepalived. Основные функции:
- Автоматический выбор MASTER/BACKUP нод
- Динамическое выделение свободного VIP
- Интеграция health-check скриптов
- Генерация безопасных паролей аутентификации

## Структура директории
```text
keepalived/
├── defaults/
│ └── main.yml # Стандартные переменные
├── files/
│ └── get_free_vip.py # Скрипт поиска свободного IP
├── handlers/
│ └── main.yml # Обработчики сервиса
├── tasks/
│ └── main.yml # Основные задачи
├── templates/
│ ├── check_service.sh.j2 # Шаблон health-check
│ └── keepalived.conf.j2 # Шаблон конфига
└── README.md # Этот файл
```


## Ключевые особенности

### 1. Автоматическое выделение VIP
Скрипт `get_free_vip.py` находит свободный IP в сети:
```bash
./get_free_vip.py 192.168.1.0/24
# Или с диапазоном:
./get_free_vip.py 192.168.1.0/24 10 20
```

## 1. Алгоритм работы:

Проверка ping

Анализ ARP-таблицы

Поиск от конца диапазона к началу

Поддержка CIDR и пользовательских диапазонов

## 2. Health-check система

Шаблон скрипта проверки сервиса:

#!/bin/bash
if systemctl is-active --quiet {{ keepalived_monitored_service }}; then
    exit 0  # сервис активен
else
    exit 1  # сервис не работает
fi

## 3. Динамическая конфигурация

Генерация конфига на основе переменных:

```jinja
vrrp_instance VI_{{ keepalived_router_id }} {
    state {{ keepalived_state }}  # MASTER/BACKUP
    interface {{ keepalived_interface }}
    priority {{ keepalived_priority }} # 110 для первого хоста, 105 для второго и т.д.
    
    authentication {
        auth_pass {{ keepalived_auth_pass }} # автогенерация
    }
}
```

## Основные задачи (tasks/main.yml)
Этапы работы:
1. Подготовка окружения:

   Создание системного пользователя keepalived_script

   Установка arping (требуется для VRRP)

2. Сетевая конфигурация:

   Автоопределение основного интерфейса
   
   Получение локального IP адреса
   
   Расчет приоритета (110, 105, 100...)

3. Выделение VIP:

   Запуск скрипта get_free_vip.py
   
   Установка router_id = последний октет VIP

4. Развертывание:

   Установка health-check скрипта
   
   Генерация конфига keepalived
   
   Запуск и включение сервиса
   
## Переменные (defaults/main.yml)
```yaml
# Базовые параметры
check_default_gw_ip: 8.8.8.8
# если не заданы, берется последний свободный ip address
# если задан только start_octet, берется последний свободный ip address, но не меньше чем start_octet
start_octet: 200
end_octet: 230

# Автоматически вычисляемые:
keepalived_interface: eth0           # Основной интерфейс
keepalived_priority: 110             # Приоритет MASTER
keepalived_state: MASTER             # Роль ноды
keepalived_virtual_ip: 192.168.1.100 # Виртуальный IP
keepalived_auth_pass: AbCdEfGh       # Автогенерируемый пароль
```

## Использование

Минимальная конфигурация:

```yaml
# inventory.ini
[keepalived_nodes]
lb1 ansible_host=192.168.1.10
lb2 ansible_host=192.168.1.11
```

## Playbook

```yaml
- hosts: keepalived_nodes
  vars:
    keepalived_monitored_service: haproxy  # Сервис для мониторинга
    keepalived_network: 192.168.1.0/24     # Сеть для VIP
  roles:
    - keepalived
```

## Расширенный health-check

   Замените шаблон check_service.sh.j2 на кастомный скрипт:

```bash
#!/bin/bash
# Проверка нескольких сервисов
if systemctl is-active nginx && curl -s http://localhost/health; then
    exit 0
fi
exit 1
```

## Требования

   Python 3 на control-ноде (для скрипта VIP)
   
   iproute2 и arping на управляемых нодах
   
   Keepalived 2.0+
   
   Systemd для health-check

## Особенности безопасности

   Скрипт VIP запускается с control-ноды
   
   Пользователь keepalived_script имеет ограниченные права
   
## Важные примечания

   Порядок хостов определяет приоритет:

```yaml
keepalived_priority: "{{ 110 - (ansible_play_hosts.index(inventory_hostname) * 5) }}"
```

   Первый хост в группе становится MASTER
   

## Для работы в облачных средах может потребоваться:

   Настройка allowed addresses
   
   Разрешение VRRP трафика
   
   Использование unicast вместо multicast
   
   Все изменения конфига вызывают перезапуск keepalived
   
