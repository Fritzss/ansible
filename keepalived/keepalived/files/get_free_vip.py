#!/usr/bin/python3

import ipaddress
import subprocess
import sys

# Получаем подсеть от аргумента командной строки
network = ipaddress.ip_network(sys.argv[1], strict=False)


def is_ip_free(ip):
    try:
        # Пинг
        ping = subprocess.run(["ping", "-c", "1", "-W", "1", str(ip)],
                              stdout=subprocess.DEVNULL,
                              stderr=subprocess.DEVNULL)
        if ping.returncode == 0:
            print(f'ping {ping.returncode}')
        # Проверка ARP (ip neigh)
            arp = subprocess.run(["ip", "neigh", "show", str(ip)],
                             stdout=subprocess.PIPE,
                             stderr=subprocess.DEVNULL)
            if b"REACHABLE" in arp.stdout or b"STALE" in arp.stdout:
               print(f'arp {arp.stdout}')
               return False

    except Exception:
        print("False")
        return False
    print("true")
    return True

# Поиск свободного IP
vip_candidate = None
for ip in network.hosts():
    if is_ip_free(ip):
        vip_candidate = str(ip)
        break

print(vip_candidate)
