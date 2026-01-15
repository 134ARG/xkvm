#!/bin/bash
# Migration script to remove network_config from existing kvm_config.json files
# Network configuration is now read-only and managed by the OS

set -e

CONFIG_FILE="/userdata/kvm_config.json"
BACKUP_FILE="/userdata/kvm_config.json.pre-network-migration"

echo "Network Config Migration Script"
echo "================================"
echo ""

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Config file not found at $CONFIG_FILE"
    echo "Nothing to migrate."
    exit 0
fi

# Check if network_config exists in the file
if ! grep -q '"network_config"' "$CONFIG_FILE"; then
    echo "No network_config found in $CONFIG_FILE"
    echo "Already migrated or never had network config."
    exit 0
fi

echo "Found network_config in $CONFIG_FILE"
echo "Creating backup at $BACKUP_FILE"

# Create backup
cp "$CONFIG_FILE" "$BACKUP_FILE"

echo "Removing network_config section..."

# Use jq to remove network_config if available, otherwise use python
if command -v jq &> /dev/null; then
    echo "Using jq for JSON manipulation"
    jq 'del(.network_config)' "$CONFIG_FILE" > "${CONFIG_FILE}.tmp"
    mv "${CONFIG_FILE}.tmp" "$CONFIG_FILE"
elif command -v python3 &> /dev/null; then
    echo "Using python3 for JSON manipulation"
    python3 << 'EOF'
import json
import sys

config_file = "/userdata/kvm_config.json"

try:
    with open(config_file, 'r') as f:
        config = json.load(f)
    
    if 'network_config' in config:
        del config['network_config']
        print(f"Removed network_config from {config_file}")
    
    with open(config_file, 'w') as f:
        json.dump(config, f, indent=2)
        f.write('\n')
    
    print("Migration completed successfully")
except Exception as e:
    print(f"Error: {e}", file=sys.stderr)
    sys.exit(1)
EOF
else
    echo "ERROR: Neither jq nor python3 found. Cannot migrate config."
    echo "Please install jq or python3 and run this script again."
    exit 1
fi

echo ""
echo "Migration completed successfully!"
echo "Backup saved at: $BACKUP_FILE"
echo ""
echo "Network configuration is now read-only and managed by your OS."
echo "Use NetworkManager, systemd-networkd, or your distribution's network tools."
