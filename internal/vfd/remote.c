#include "remote.h"

#include <pthread.h>
#include <stdbool.h>
#include <string.h>
#include <time.h>

typedef struct {
    metrics_snapshot_t snapshot;
    time_t last_update;
    pthread_mutex_t lock;
    bool connected;
} vfd_metrics_t;

static vfd_metrics_t vfd_metrics = {
    .lock = PTHREAD_MUTEX_INITIALIZER,
};

void vfd_set_remote_metrics(const metrics_snapshot_t* in, bool connected) {
    pthread_mutex_lock(&vfd_metrics.lock);
    if (in) {
        vfd_metrics.snapshot = *in;
        vfd_metrics.last_update = time(NULL);
    } else {
        memset(&vfd_metrics.snapshot, 0, sizeof(vfd_metrics.snapshot));
    }
    vfd_metrics.connected = connected;
    pthread_mutex_unlock(&vfd_metrics.lock);
}

void get_remote_metrics_full(metrics_snapshot_t* out) {
    pthread_mutex_lock(&vfd_metrics.lock);
    *out = vfd_metrics.snapshot;
    pthread_mutex_unlock(&vfd_metrics.lock);
}

void get_remote_metrics(double* cpu, double* ram, double* gpu) {
    pthread_mutex_lock(&vfd_metrics.lock);
    *cpu = vfd_metrics.snapshot.cpu_util;
    *ram = vfd_metrics.snapshot.ram_util;
    *gpu = vfd_metrics.snapshot.gpu_util;
    pthread_mutex_unlock(&vfd_metrics.lock);
}

bool is_remote_connected(void) {
    pthread_mutex_lock(&vfd_metrics.lock);
    bool c = vfd_metrics.connected;
    pthread_mutex_unlock(&vfd_metrics.lock);
    return c;
}
