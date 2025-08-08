# System Tuning Role

Эта роль выполняет оптимизацию параметров операционной системы для production-сред. Основные направления:
- Настройка параметров ядра (sysctl)
- Установка лимитов ресурсов (limits.conf)
- Требует перезагрузки для применения изменений

## Структура директории
```text
tuning/
├── files/
│ └── limits.conf # Шаблон лимитов ресурсов
├── handlers/
│ └── main.yml # Обработчики перезагрузки
├── tasks/
│ └── main.yml # Основные задачи настройки
└── README.md # Этот файл
```

## Основные компоненты

### 1. Настройка sysctl-параметров (tasks/main.yml)
Оптимизирует системные параметры ядра:

```yaml
- sysctl:
    name: vm.swappiness
    value: 0  # Отключает свопинг когда возможно

- sysctl:
    name: vm.dirty_ratio
    value: 80  # Максимум "грязных" страниц в памяти

- sysctl:
    name: vm.dirty_background_ratio
    value: 5  # Фоновая запись "грязных" страниц

- sysctl:
    name: net.core.wmem_max
    value: 16777216  # Макс. размер буфера записи

# ...и другие критические параметры
```
## Полный список настроек:

Параметры виртуальной памяти (vm.*)

Сетевые буферы (net.core.*)

TCP-параметры (net.ipv4.tcp_*)

Лимиты файловых дескрипторов (fs.file-max)

## 2. Установка лимитов ресурсов (files/limits.conf)

```conf
* soft nofile 1280000    # Макс. открытых файлов (мягкий лимит)
* hard nofile 1280000    # Макс. открытых файлов (жесткий лимит)
* soft nproc 65536       # Макс. процессов (мягкий)
* hard nproc 65536       # Макс. процессов (жесткий)
* soft memlock unlimited # Блокировка памяти
* hard memlock unlimited
* soft as unlimited      # Виртуальная память
* hard as unlimited
```
## 3. Обработчик перезагрузки (handlers/main.yml)

```yaml
- name: reboot_host
  reboot:
    msg: "Applying kernel tuning changes"
    connect_timeout: 10
    test_command: uptime
```


## Особенности реализации
Атомарная настройка параметров
Каждый параметр настраивается отдельным модулем sysctl для точного контроля.

Автоматическая перезагрузка
Все изменения помечены notify: reboot_host - хендлер выполняется один раз в конце плейбука.

Безопасное применение лимитов
Файл limits.conf полностью заменяется на оптимизированную версию:
```yaml
- name: Apply resource limits
  copy:
    src: files/limits.conf
    dest: /etc/security/limits.conf
    force: yes
```

## Важные замечания
# 1. Процесс перезагрузки
Перезагрузка выполняется через модуль reboot

Проверка успешности: команда uptime после перезагрузки

Таймаут подключения: 10 секунд

# 2. Закомментированные альтернативы
В задачах присутствует альтернативный подход к настройке лимитов через lineinfile, который:

Удаляет существующие настройки перед применением новых

Требует точного указания параметров

В текущей реализации не используется

# 3. Критические параметры
Параметр	Значение	Назначение
vm.swappiness	0	Минимизация свопинга
vm.max_map_count	262144	Для memory-intensive БД
fs.file-max	1048576	Макс. открытых файлов в системе
net.core.wmem_max	16777216	16MB буфер записи TCP
net.ipv4.tcp_max_syn_backlog	4096	Очередь SYN-запросов

# Использование
Добавьте в ваш плейбук:
```yaml
- hosts: all
  roles:
    - tuning
```

# Роль автоматически:

Применит все настройки ядра

Обновит /etc/security/limits.conf

Выполнит перезагрузку с проверкой

# Требования
Ansible 2.9+

Права root/sudo

Доступ к модулю reboot (обычно требует become)

# Безопасность
Проверяйте параметры memlock unlimited - может быть нежелательно в multi-user средах

Значение nproc 65536 может потребовать настройки systemd для сервисов

Все изменения требуют перезагрузки - планируйте downtime

# Кастомизация
Для изменения параметров создайте переменные в group_vars или host_vars:
```yaml
# group_vars/all/tuning.yml
sysctl_params:
  vm.swappiness: 1
  net.core.wmem_max: 33554432
  fs.file-max: 2097152

custom_limits: |
  * soft nofile 256000
  * hard nofile 512000
  application soft memlock unlimited
  application hard memlock unlimited
```
