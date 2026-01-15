# Embedded Libraries - Self-Extracting Installer

## Overview

The build system creates a self-extracting installer (`xkvm_installer.sh`) that packages the XKVM application with all required vendor libraries. This eliminates manual library deployment and dependency issues.

## Problem

The application depends on vendor-provided dynamic libraries:
- `librockit.so` (2.3MB) - Rockchip video encoding library  
- `libgraphic_lsf.so` (164KB) - Graphics layer composition library
- `librockchip_mpp.so*` - Media Process Platform libraries

These libraries are:
- Closed source
- Binary-only distribution
- Required at application startup (before Go code runs)

## Solution

A self-extracting shell script that:
1. Contains the binary and all libraries in a compressed tar archive
2. Extracts everything to the target directory with proper structure
3. Preserves the directory hierarchy needed by RPATH

## Usage

### Building

```bash
ARM64_SYSROOT=/path/to/sysroot make build_release
```

This creates:
- `bin/xkvm_app` - The main executable (with embedded library data)
- `bin/xkvm_installer.sh` - Self-extracting installer (~19MB)

### Installing on Target Device

**Basic installation (extract to current directory):**
```bash
./xkvm_installer.sh
```

**Install to specific directory:**
```bash
./xkvm_installer.sh --dir /opt/xkvm
```

**Skip extraction if files exist:**
```bash
./xkvm_installer.sh --skip-extract
```

**Show help:**
```bash
./xkvm_installer.sh --help
```

### Running

After extraction:
```bash
cd /path/to/extracted/directory
sudo ./xkvm_app
```

## Directory Structure After Extraction

```
.
├── xkvm_app              # Main executable
└── lib/                    # Shared libraries
    ├── librockit.so
    ├── libgraphic_lsf.so
    └── librockchip_mpp.so*
```

## How It Works

### Build Process

1. **Compile Binary**: Go binary is built with RPATH set to `$ORIGIN/lib`
2. **Collect Libraries**: Script gathers all required `.so` files
3. **Create Archive**: Files are packaged into a tar.gz
4. **Generate Installer**: Shell script + archive = self-extracting installer

### Runtime

1. **Extraction**: User runs installer, files extracted with hierarchy preserved
2. **Dynamic Linking**: When `xkvm_app` starts, dynamic linker uses RPATH
3. **Library Loading**: Finds libraries in `./lib/` relative to executable

### RPATH Configuration

The binary is built with:
```
-extldflags '-Wl,-rpath,\$ORIGIN/lib'
```

This tells the dynamic linker to search for libraries in the `lib` directory relative to the executable's location.

## Implementation Details

### Files

1. **scripts/create_self_extract.sh** - Creates the self-extracting installer
2. **Makefile** - Updated `_build_release_inner` target to call the script
3. **internal/native/embedded_libs.go** - Go embed code (kept for future use)

### Self-Extracting Script Structure

```bash
#!/bin/bash
# ... extraction logic ...
__ARCHIVE_BELOW__
<binary tar.gz data>
```

The script:
- Parses command-line arguments
- Finds the `__ARCHIVE_BELOW__` marker
- Extracts everything after the marker using `tail` and `tar`

## Advantages

1. **Single File Distribution**: One file contains everything
2. **No Manual Steps**: Extraction is automatic
3. **Flexible Deployment**: Can extract to any directory
4. **Standard Tools**: Uses only `bash`, `tar`, `gzip` (available everywhere)
5. **Idempotent**: Can skip extraction if files exist
6. **Portable**: Works on any Linux ARM64 system

## Testing

Verify the installer:
```bash
# Check file size
ls -lh bin/xkvm_installer.sh

# Test extraction to temp directory
mkdir /tmp/test-install
./bin/xkvm_installer.sh --dir /tmp/test-install

# Verify structure
ls -la /tmp/test-install
ls -la /tmp/test-install/lib

# Check RPATH
readelf -d /tmp/test-install/xkvm_app | grep RPATH

# Test execution
cd /tmp/test-install
sudo ./xkvm_app
```

## Deployment Workflow

### Development
```bash
# Build
ARM64_SYSROOT=/path/to/sysroot make build_release

# Copy installer to device
scp bin/xkvm_installer.sh user@device:/tmp/

# On device
ssh user@device
cd /opt
sudo /tmp/xkvm_installer.sh --dir .
sudo ./xkvm_app
```

### Production
```bash
# Upload to release server
./bin/xkvm_installer.sh

# Users download and run
wget https://example.com/xkvm_installer.sh
chmod +x xkvm_installer.sh
sudo ./xkvm_installer.sh --dir /opt/xkvm
```

## Troubleshooting

### "cannot open shared object file"
- Ensure extraction completed successfully
- Check that `lib/` directory exists next to `xkvm_app`
- Verify RPATH: `readelf -d xkvm_app | grep RPATH`

### "Permission denied"
- Make sure installer is executable: `chmod +x xkvm_installer.sh`
- May need sudo for extraction to system directories

### Libraries not found
- Check library files exist: `ls -la lib/`
- Verify library permissions: `chmod 755 lib/*.so`

## Future Improvements

- Add checksum verification
- Support for incremental updates
- Compression options (xz, bzip2)
- Digital signature verification
- Automatic cleanup of old versions
