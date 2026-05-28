#!/usr/bin/env python3
import argparse
import glob
import json
import os
import socket
import subprocess
import time

import psutil

BOOT_TIME = psutil.boot_time()
AMD_GPU_BUSY_PATH = None
AMD_GPU_TEMP_PATH = None
CPU_TEMP_PATH = None
LAST_SYSTEMD_CHECK = 0.0
CACHED_FAILED_UNITS = 0
CONNECT_TIMEOUT_SECONDS = 2.0
IGNORED_NET_PREFIXES = (
    "lo",
    "docker",
    "br-",
    "veth",
    "vnet",
    "vmnet",
    "virbr",
    "lxc",
    "lxdbr",
    "podman",
    "cni",
    "flannel",
    "kube",
    "tunl",
    "tailscale",
    "zt",
    "zerotier",
    "wg",
    "tun",
    "tap",
    "ppp",
    "ipsec",
)


def init_hardware_paths():
    global AMD_GPU_BUSY_PATH, AMD_GPU_TEMP_PATH, CPU_TEMP_PATH

    paths = glob.glob("/sys/class/drm/card*/device/gpu_busy_percent")
    if paths:
        AMD_GPU_BUSY_PATH = paths[0]
        print(f"gpu busy path: {AMD_GPU_BUSY_PATH}")

    for hwmon in glob.glob("/sys/class/drm/card*/device/hwmon/hwmon*"):
        for i in range(1, 10):
            label_path = f"{hwmon}/temp{i}_label"
            input_path = f"{hwmon}/temp{i}_input"
            try:
                with open(label_path) as f:
                    if f.read().strip() == "junction" and os.path.exists(input_path):
                        AMD_GPU_TEMP_PATH = input_path
                        print(f"gpu temp path: {AMD_GPU_TEMP_PATH}")
                        break
            except FileNotFoundError:
                pass
        if AMD_GPU_TEMP_PATH:
            break

    for hwmon in glob.glob("/sys/class/hwmon/hwmon*"):
        for i in range(1, 10):
            label_path = f"{hwmon}/temp{i}_label"
            input_path = f"{hwmon}/temp{i}_input"
            try:
                with open(label_path) as f:
                    if f.read().strip() == "Tctl" and os.path.exists(input_path):
                        CPU_TEMP_PATH = input_path
                        print(f"cpu temp path: {CPU_TEMP_PATH}")
                        break
            except FileNotFoundError:
                pass
        if CPU_TEMP_PATH:
            break


def read_int(path, scale=1):
    try:
        with open(path, "rb", buffering=0) as f:
            return int(f.read().strip()) / scale
    except Exception:
        return 0


def read_float(path):
    try:
        with open(path, "rb", buffering=0) as f:
            return float(f.read().strip())
    except Exception:
        return 0.0


def collect_failed_units(now):
    global LAST_SYSTEMD_CHECK, CACHED_FAILED_UNITS

    if now - LAST_SYSTEMD_CHECK <= 20.0:
        return CACHED_FAILED_UNITS

    try:
        result = subprocess.run(
            ["systemctl", "--no-legend", "--plain", "--state=failed", "list-units"],
            capture_output=True,
            text=True,
            timeout=2,
        )
        CACHED_FAILED_UNITS = sum(
            1 for line in result.stdout.splitlines() if line.strip()
        )
        LAST_SYSTEMD_CHECK = now
    except Exception:
        pass
    return CACHED_FAILED_UNITS


def collect_net_bytes():
    rx_bytes = 0
    tx_bytes = 0

    try:
        for iface, counters in psutil.net_io_counters(pernic=True).items():
            if iface.startswith(IGNORED_NET_PREFIXES):
                continue
            rx_bytes += counters.bytes_recv
            tx_bytes += counters.bytes_sent
        return rx_bytes, tx_bytes
    except Exception:
        counters = psutil.net_io_counters()
        return counters.bytes_recv, counters.bytes_sent


def collect_metrics():
    now = time.time()
    mem = psutil.virtual_memory()
    net_rx_bytes, net_tx_bytes = collect_net_bytes()

    gpu_util = read_float(AMD_GPU_BUSY_PATH) if AMD_GPU_BUSY_PATH else 0.0
    gpu_temp = int(read_int(AMD_GPU_TEMP_PATH, 1000)) if AMD_GPU_TEMP_PATH else 0
    cpu_temp = int(read_int(CPU_TEMP_PATH, 1000)) if CPU_TEMP_PATH else 0

    return {
        "ts": int(now),
        "metrics": {
            "cpu": {"util": round(psutil.cpu_percent(interval=None), 1)},
            "ram": {
                "util": round(mem.percent, 1),
                "total_mb": mem.total // (1024 * 1024),
                "used_mb": mem.used // (1024 * 1024),
            },
            "gpu": {"util": round(gpu_util, 1), "temp": gpu_temp},
            "temp": {"cpu": cpu_temp, "gpu": gpu_temp},
            "net": {"rx_bytes": net_rx_bytes, "tx_bytes": net_tx_bytes},
            "sys": {
                "uptime": int(now - BOOT_TIME),
                "failed_units": collect_failed_units(now),
            },
        },
    }


def push_loop(host, port, interval):
    psutil.cpu_percent(interval=None)
    init_hardware_paths()

    while True:
        sock = None
        try:
            print(f"connecting to {host}:{port}")
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(CONNECT_TIMEOUT_SECONDS)
            sock.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
            sock.connect((host, port))
            sock.settimeout(None)
            print(f"connected to {host}:{port}")

            while True:
                line = json.dumps(collect_metrics(), separators=(",", ":")) + "\n"
                sock.sendall(line.encode())
                time.sleep(interval)
        except (ConnectionRefusedError, ConnectionResetError, BrokenPipeError, OSError) as e:
            print(f"connection lost ({e}), retrying in 2s")
            if sock:
                try:
                    sock.close()
                except Exception:
                    pass
            time.sleep(2)


def main():
    parser = argparse.ArgumentParser(description="Metric agent for xkVM VFD display")
    parser.add_argument("--host", required=True, help="xkVM VFD listener host")
    parser.add_argument("--port", type=int, default=9101, help="xkVM VFD listener port")
    parser.add_argument("--interval", type=float, default=0.25, help="push interval")
    args = parser.parse_args()
    push_loop(args.host, args.port, args.interval)


if __name__ == "__main__":
    main()
