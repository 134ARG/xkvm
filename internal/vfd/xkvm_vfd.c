#include "xkvm_vfd.h"

#include <stdatomic.h>
#include <pthread.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include "remote.h"
#include "spi.h"
#include "themes.h"
#include "vfd.h"

static pthread_t render_thread;
static atomic_bool render_running = false;

#define VFD_RENDER_INTERVAL_US (10 * 1000)

static void* render_loop(void* arg) {
    (void)arg;
    while (atomic_load_explicit(&render_running, memory_order_acquire)) {
        numeric_map_monitor();
        usleep(VFD_RENDER_INTERVAL_US);
    }
    return NULL;
}

int xkvm_vfd_init(const char* device_path) {
    if (atomic_load_explicit(&render_running, memory_order_acquire))
        return 0;

    ch347_setup_signal();
    spi = &ch347_backend;

    char* detected_path = NULL;
    if (!device_path || device_path[0] == '\0') {
        detected_path = ch347_auto_detect();
        device_path = detected_path;
    }
    if (!device_path) {
        fprintf(stderr, "VFD: CH347 auto-detect failed\n");
        return -1;
    }

    if (!spi->init(device_path)) {
        free(detected_path);
        return -2;
    }
    free(detected_path);

    vfd_init();
    sleep(1);

    atomic_store_explicit(&render_running, true, memory_order_release);
    if (pthread_create(&render_thread, NULL, render_loop, NULL) != 0) {
        atomic_store_explicit(&render_running, false, memory_order_release);
        spi->close();
        return -3;
    }
    return 0;
}

void xkvm_vfd_shutdown(void) {
    if (!atomic_exchange_explicit(&render_running, false, memory_order_acq_rel))
        return;

    pthread_join(render_thread, NULL);
    if (spi) {
        spi->close();
    }
}

void xkvm_vfd_update_host_metrics(const xkvm_vfd_host_metrics_t* metrics) {
    if (!metrics) {
        vfd_set_remote_metrics(NULL, false);
        return;
    }

    metrics_snapshot_t snapshot = {
        .cpu_util = metrics->cpu_util,
        .ram_util = metrics->ram_util,
        .gpu_util = metrics->gpu_util,
        .cpu_temp = metrics->cpu_temp,
        .gpu_temp = metrics->gpu_temp,
        .net_rx_bytes = metrics->net_rx_bytes,
        .net_tx_bytes = metrics->net_tx_bytes,
        .uptime_sec = metrics->uptime_sec,
        .failed_units = metrics->failed_units,
    };
    vfd_set_remote_metrics(&snapshot, metrics->connected);
}
