#!/bin/bash
set -e

DEB="$1"
if [ -z "$DEB" ] || [ ! -f "$DEB" ]; then
    echo "usage: $0 <path-to-deb>" >&2
    exit 1
fi

# Brief delay so the calling RPC can return before the service restarts.
sleep 1

dpkg -i "$DEB"
systemctl restart xkvm.service
rm -f "$DEB"
