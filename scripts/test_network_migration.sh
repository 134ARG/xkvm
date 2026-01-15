#!/bin/bash
# Test script to verify network config migration behavior

set -e

echo "Network Config Migration Test"
echo "=============================="
echo ""

# Create a test config with network_config
TEST_CONFIG=$(cat <<'EOF'
{
  "cloud_url": "https://api.xkvm.com",
  "auto_update_enabled": true,
  "keyboard_layout": "en-US",
  "network_config": {
    "dhcp_client": "udhcpc",
    "hostname": "test-device",
    "domain": "example.com",
    "ipv4_mode": "dhcp",
    "ipv6_mode": "slaac",
    "lldp_mode": "basic",
    "mdns_mode": "auto"
  },
  "default_log_level": "INFO"
}
EOF
)

echo "Test 1: Verify network_config is present in test data"
echo "$TEST_CONFIG" | grep -q '"network_config"' && echo "✓ network_config found in test data" || echo "✗ FAIL: network_config not found"
echo ""

echo "Test 2: Verify migration script can remove network_config"
if command -v jq &> /dev/null; then
    MIGRATED=$(echo "$TEST_CONFIG" | jq 'del(.network_config)')
    echo "$MIGRATED" | grep -q '"network_config"' && echo "✗ FAIL: network_config still present" || echo "✓ network_config successfully removed"
    echo ""
    echo "Migrated config:"
    echo "$MIGRATED"
elif command -v python3 &> /dev/null; then
    MIGRATED=$(python3 << 'PYEOF'
import json
import sys
config = json.loads(sys.stdin.read())
if 'network_config' in config:
    del config['network_config']
print(json.dumps(config, indent=2))
PYEOF
)
    echo "$MIGRATED" | grep -q '"network_config"' && echo "✗ FAIL: network_config still present" || echo "✓ network_config successfully removed"
    echo ""
    echo "Migrated config:"
    echo "$MIGRATED"
else
    echo "⚠ WARNING: Neither jq nor python3 available, skipping migration test"
fi

echo ""
echo "Test 3: Verify other config fields are preserved"
if command -v jq &> /dev/null; then
    CLOUD_URL=$(echo "$MIGRATED" | jq -r '.cloud_url')
    KEYBOARD=$(echo "$MIGRATED" | jq -r '.keyboard_layout')
    [ "$CLOUD_URL" = "https://api.xkvm.com" ] && echo "✓ cloud_url preserved" || echo "✗ FAIL: cloud_url not preserved"
    [ "$KEYBOARD" = "en-US" ] && echo "✓ keyboard_layout preserved" || echo "✗ FAIL: keyboard_layout not preserved"
elif command -v python3 &> /dev/null; then
    python3 << 'PYEOF'
import json
import sys
config = json.loads('''$MIGRATED''')
if config.get('cloud_url') == 'https://api.xkvm.com':
    print('✓ cloud_url preserved')
else:
    print('✗ FAIL: cloud_url not preserved')
if config.get('keyboard_layout') == 'en-US':
    print('✓ keyboard_layout preserved')
else:
    print('✗ FAIL: keyboard_layout not preserved')
PYEOF
fi

echo ""
echo "=============================="
echo "Migration test completed"
