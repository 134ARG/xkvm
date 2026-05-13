#pragma once

#include <stdbool.h>

typedef struct {
    double cpu_util;
    double ram_util;
    double gpu_util;
    double cpu_temp;
    double gpu_temp;
    double net_rx_bytes;
    double net_tx_bytes;
    int uptime_sec;
    int failed_units;
    bool connected;
} xkvm_vfd_host_metrics_t;

int xkvm_vfd_init(const char* device_path);
void xkvm_vfd_shutdown(void);
void xkvm_vfd_update_host_metrics(const xkvm_vfd_host_metrics_t* metrics);
