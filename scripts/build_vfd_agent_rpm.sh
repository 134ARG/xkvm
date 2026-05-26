#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

VERSION="$(tr -d '[:space:]' < packaging/version.txt)"
PKG_NAME="xkvm-vfd-agent"
TOPDIR="$PROJECT_ROOT/build/vfd-agent-rpmbuild"
SOURCE="$TOPDIR/SOURCES/$PKG_NAME-$VERSION.tar.gz"
TMPDIR="$TOPDIR/tmp"

if ! command -v rpmbuild >/dev/null 2>&1; then
    echo "rpmbuild not found"
    exit 1
fi

mkdir -p "$TOPDIR"/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS,tmp} bin
rm -f "$SOURCE"

tar -czf "$SOURCE" \
    --exclude='__pycache__' \
    --exclude='*.pyc' \
    --transform "s,^,$PKG_NAME-$VERSION/," \
    agents/vfd packaging/vfd-agent

rpmbuild -bb packaging/vfd-agent/rpm/xkvm-vfd-agent.spec \
    --define "_topdir $TOPDIR" \
    --define "_tmppath $TMPDIR" \
    --define "xkvm_version $VERSION"

RPM_FILE="$(find "$TOPDIR/RPMS" -type f -name "$PKG_NAME-$VERSION-*.rpm" | head -n 1)"
if [ -z "$RPM_FILE" ]; then
    echo "RPM build finished but output package was not found"
    exit 1
fi

cp "$RPM_FILE" "bin/"
echo "RPM package created: bin/$(basename "$RPM_FILE")"
