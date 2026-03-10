#!/bin/bash
export LD_LIBRARY_PATH=/usr/lib/xkvm:$LD_LIBRARY_PATH
exec /usr/lib/xkvm/xkvm_app "$@"
