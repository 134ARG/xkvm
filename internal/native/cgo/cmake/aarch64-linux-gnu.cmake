# CMake toolchain file for cross-compiling to ARM64/AArch64
# Usage: cmake -DCMAKE_TOOLCHAIN_FILE=cmake/aarch64-linux-gnu.cmake

set(CMAKE_SYSTEM_NAME Linux)
set(CMAKE_SYSTEM_PROCESSOR aarch64)

# Set the sysroot path from environment variable
if(DEFINED ENV{ARM64_SYSROOT})
    set(CMAKE_SYSROOT $ENV{ARM64_SYSROOT})
    message(STATUS "Using ARM64 sysroot: ${CMAKE_SYSROOT}")
else()
    message(FATAL_ERROR "ARM64_SYSROOT environment variable is required for cross-compilation")
endif()

# Cross-compilation tools
set(CMAKE_C_COMPILER aarch64-linux-gnu-gcc)
set(CMAKE_CXX_COMPILER aarch64-linux-gnu-g++)
set(CMAKE_AR aarch64-linux-gnu-ar)
set(CMAKE_STRIP aarch64-linux-gnu-strip)
set(CMAKE_RANLIB aarch64-linux-gnu-ranlib)

# Set the find root path to our sysroot
set(CMAKE_FIND_ROOT_PATH ${CMAKE_SYSROOT})

# Search for programs in the build host directories
set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)
# Search for libraries and headers in the target directories
set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE ONLY)

# Set cross-compilation flag
set(CMAKE_CROSSCOMPILING TRUE)

# Skip compiler tests that require linking (since we have a working cross-compiler)
set(CMAKE_C_COMPILER_WORKS TRUE)
set(CMAKE_CXX_COMPILER_WORKS TRUE)

# Compiler flags for ARM64 Cortex-A55 (RK3566) with sysroot and include paths
set(CMAKE_C_FLAGS_INIT "-march=armv8.2-a+fp16 -mtune=cortex-a55 --sysroot=${CMAKE_SYSROOT} -I${CMAKE_SYSROOT}/usr/include/aarch64-linux-gnu")
set(CMAKE_CXX_FLAGS_INIT "-march=armv8.2-a+fp16 -mtune=cortex-a55 --sysroot=${CMAKE_SYSROOT} -I${CMAKE_SYSROOT}/usr/include/aarch64-linux-gnu")

# Linker flags - add lib64 directory for startup files and libraries
set(CMAKE_EXE_LINKER_FLAGS_INIT "--sysroot=${CMAKE_SYSROOT} -L${CMAKE_SYSROOT}/lib64 -Wl,-rpath-link,${CMAKE_SYSROOT}/lib64")
set(CMAKE_SHARED_LINKER_FLAGS_INIT "--sysroot=${CMAKE_SYSROOT} -L${CMAKE_SYSROOT}/lib64 -Wl,-rpath-link,${CMAKE_SYSROOT}/lib64")

# Tell the compiler where to find startup files
set(CMAKE_C_LINK_FLAGS "-B${CMAKE_SYSROOT}/lib64")
set(CMAKE_CXX_LINK_FLAGS "-B${CMAKE_SYSROOT}/lib64")

# Add library search paths
link_directories(${CMAKE_SYSROOT}/lib64)
link_directories(${CMAKE_SYSROOT}/usr/lib/aarch64-linux-gnu)

# Enable building the binary
set(BUILD_BINARY ON CACHE BOOL "Build binary executable" FORCE)
