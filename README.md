# ansible
# Ansible Infrastructure Automation Repository

Этот репозиторий содержит набор Ansible-ролей для настройки компонентов инфраструктуры.

## Содержание ролей

1. **`SSL`**  
   Автоматизация развёртывания и обновления SSL/TLS сертификатов:
   - Генерация self-signed сертификатов
   - Поддержка SAN

2. **`tuning`**  
   Оптимизация параметров операционной системы для production-нагрузки:
   - Настройка sysctl-параметров (сетевой стек, VM)
   - Оптимизация limits.conf
   - Файловая система и дисковая подсистема

3. **`keepalived`**  
   Настройка отказоустойчивого кластера с использованием VRPP:
   - Виртуальные IP-адреса (VIP)
   - Health-check скрипты
   - Конфигурация для master/backup нод
   - Интеграция с service

4. **`simple-clickhouse-cluster`**  
   Развёртывание кластера ClickHouse:
   - Установка и настройка ClickHouse Server cluster репликации и шардирование
   - Конфигурация Clickhouse-keeper


## Требования
- Ansible ≥ 2.12
- Доступ с правами root/sudo
