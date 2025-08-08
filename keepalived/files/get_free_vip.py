#!/usr/bin/python3

import ipaddress
import subprocess
import sys

def print_usage():
    print("Usage:")
    print(f"  {sys.argv[0]} <network> [<start_octet> <end_octet>]")
    print("\nExamples:")
    print(f"  {sys.argv[0]} 192.168.1.0/24")
    print(f"  {sys.argv[0]} 192.168.1.0/24 10 20")
    print("\nNote:")
    print("  - Start and end octets must be between 1-254")
    print("  - Octets must be within the network range")
    sys.exit(1)

# Проверка аргументов
if len(sys.argv) not in [2,3,4]:
    print("Error: Invalid number of arguments")
    print_usage()

# Получаем подсеть
try:
    network = ipaddress.ip_network(sys.argv[1], strict=False)
except ValueError as e:
    print(f"Invalid network format: {e}")
    print_usage()

# Функция проверки свободного IP
def is_ip_free(ip):
    """Проверяет, свободен ли IP-адрес."""
    try:
        # Проверка с помощью ping
        ping = subprocess.run(
            ["ping", "-c", "1", "-W", "0.3", str(ip)],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL
        )
        if ping.returncode == 0:
            return False

        # Проверка ARP таблицы
        arp = subprocess.run(
            ["ip", "neigh", "show", str(ip)],
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True
        )
        if arp.stdout and ("REACHABLE" in arp.stdout or "STALE" in arp.stdout):
            return False

    except Exception:
        return False
    return True

# Определяем диапазон поиска
if len(sys.argv) == 4:
    # Режим с указанием последних октетов
    try:
        start_octet = int(sys.argv[2])
        end_octet = int(sys.argv[3])

        # Проверка диапазона октетов
        if not (1 <= start_octet <= 254) or not (1 <= end_octet <= 254):
            print("Error: Octets must be between 1-254")
            print_usage()

        if start_octet > end_octet:
            print("Error: Start octet must be <= end octet")
            print_usage()

        # Формируем IP-адреса из октетов
        base_ip = str(network.network_address).rsplit('.', 1)[0]
        start_ip = ipaddress.IPv4Address(f"{base_ip}.{start_octet}")
        end_ip = ipaddress.IPv4Address(f"{base_ip}.{end_octet}")

        # Проверка принадлежности к сети
        if start_ip not in network or end_ip not in network:
            # Определяем допустимый диапазон
            first_host = next(network.hosts())
            last_host = list(network.hosts())[-1]
            first_octet = int(str(first_host).rsplit('.', 1)[-1])
            last_octet = int(str(last_host).rsplit('.', 1)[-1])

            print(f"Error: Octets out of network range ({first_octet}-{last_octet})")
            print(f"First usable IP: {first_host}")
            print(f"Last usable IP: {last_host}")
            sys.exit(1)

        # Генерируем диапазон IP
        ip_range = []
        current_ip = start_ip
        while current_ip <= end_ip:
            ip_range.append(current_ip)
            current_ip += 1

        # Переворачиваем для поиска от конца к началу
        ip_range = ip_range[::-1]

    except ValueError:
        print("Error: Octets must be integers")
        print_usage()
elif len(sys.argv) == 3:
    # Режим с указанием последних октетов
    try:
        start_octet = int(sys.argv[2])

        # Проверка диапазона октетов
        if not (1 <= start_octet <= 254):
            print("Error: Octets must be between 1-254")
            print_usage()


        # Формируем IP-адреса из октетов
        last_host = list(network.hosts())[-1]
        end_octet = int(str(last_host).rsplit('.', 1)[-1])
        base_ip = str(network.network_address).rsplit('.', 1)[0]
        start_ip = ipaddress.IPv4Address(f"{base_ip}.{start_octet}")
        end_ip = ipaddress.IPv4Address(f"{base_ip}.{end_octet}")

        # Проверка принадлежности к сети
        if start_ip not in network or end_ip not in network:
            # Определяем допустимый диапазон
            first_host = next(network.hosts())
            first_octet = int(str(first_host).rsplit('.', 1)[-1])

            print(f"Error: Octets out of network range ({first_octet}-{end_octet})")
            print(f"First usable IP: {first_host}")
            print(f"Last usable IP: {last_host}")
            sys.exit(1)

        # Генерируем диапазон IP
        ip_range = []
        current_ip = start_ip
        while current_ip <= end_ip:
            ip_range.append(current_ip)
            current_ip += 1

        # Переворачиваем для поиска от конца к началу
        ip_range = ip_range[::-1]

    except ValueError:
        print("Error: Octets must be integers")
        print_usage()


else:
    # Режим без диапазона - вся сеть
    ip_range = list(network.hosts())[::-1]

# Поиск свободного IP
vip_candidate = None
for ip in ip_range:
    if is_ip_free(ip):
        vip_candidate = str(ip)
        break

print(vip_candidate if vip_candidate else "No free IP found")
