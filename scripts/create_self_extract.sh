#!/bin/bash
set -e

SCRIPT_DIR=$(dirname "$(realpath "$0")")
PROJECT_ROOT=$(realpath "$SCRIPT_DIR/..")

cd "$PROJECT_ROOT"

# Check if xkvm_app exists
if [ ! -f "bin/xkvm_app" ]; then
    echo "Error: bin/xkvm_app not found. Run 'make build_release' first."
    exit 1
fi

echo "Creating self-extracting archive..."

# Create temporary directory for packaging
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Create directory structure
mkdir -p "$TEMP_DIR/lib"

# Copy binary
cp bin/xkvm_app "$TEMP_DIR/"

# Copy vendor libraries
cp internal/native/cgo/sdk/vendor/rockit/lib/lib64/librockit.so "$TEMP_DIR/lib/"
cp internal/native/cgo/sdk/vendor/rockit/lib/lib64/libgraphic_lsf.so "$TEMP_DIR/lib/"

# Copy MPP libraries
cp internal/native/cgo/sdk/mpp/lib/librockchip_mpp.so* "$TEMP_DIR/lib/" 2>/dev/null || true

echo "Packaging files..."
cd "$TEMP_DIR"
tar czf ../payload.tar.gz .
cd ..

PAYLOAD_FILE="$TEMP_DIR/../payload.tar.gz"
OUTPUT_FILE="$PROJECT_ROOT/bin/xkvm_installer.sh"

# Create self-extracting script
cat > "$OUTPUT_FILE" << 'EOF'
#!/bin/bash
# XKVM Self-Extracting Installer
# This script extracts the XKVM application and its dependencies

set -e

EXTRACT_DIR="."
SKIP_EXTRACT=0

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dir)
            EXTRACT_DIR="$2"
            shift 2
            ;;
        --skip-extract)
            SKIP_EXTRACT=1
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --dir DIR          Extract to specified directory (default: current directory)"
            echo "  --skip-extract     Skip extraction if files already exist"
            echo "  --help             Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Check if already extracted
if [ "$SKIP_EXTRACT" -eq 1 ] && [ -f "$EXTRACT_DIR/xkvm_app" ] && [ -d "$EXTRACT_DIR/lib" ]; then
    echo "Files already extracted, skipping extraction."
    exit 0
fi

echo "XKVM Self-Extracting Installer"
echo "================================="
echo "Extracting to: $EXTRACT_DIR"
echo ""

# Create extraction directory
mkdir -p "$EXTRACT_DIR"

# Find the start of the tar archive
ARCHIVE_LINE=$(awk '/^__ARCHIVE_BELOW__/ {print NR + 1; exit 0; }' "$0")

# Extract the archive
tail -n +$ARCHIVE_LINE "$0" | tar xzf - -C "$EXTRACT_DIR"

echo ""
echo "✓ Extraction complete!"
echo ""
echo "Files extracted:"
echo "  - xkvm_app (main executable)"
echo "  - lib/librockit.so"
echo "  - lib/libgraphic_lsf.so"
echo "  - lib/librockchip_mpp.so*"
echo ""
echo "To run XKVM:"
echo "  cd $EXTRACT_DIR"
echo "  sudo ./xkvm_app"
echo ""

exit 0

__ARCHIVE_BELOW__
EOF

# Append the tar archive
cat "$PAYLOAD_FILE" >> "$OUTPUT_FILE"

# Make it executable
chmod +x "$OUTPUT_FILE"

# Cleanup
rm "$PAYLOAD_FILE"

echo ""
echo "✓ Self-extracting installer created: bin/xkvm_installer.sh"
echo ""
echo "Usage:"
echo "  ./bin/xkvm_installer.sh              # Extract to current directory"
echo "  ./bin/xkvm_installer.sh --dir /opt   # Extract to /opt"
echo "  ./bin/xkvm_installer.sh --help       # Show help"
echo ""
echo "File size: $(du -h "$OUTPUT_FILE" | cut -f1)"
