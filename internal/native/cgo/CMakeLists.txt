cmake_minimum_required(VERSION 3.14)
include(FetchContent)
include(ExternalProject)

project(jknative LANGUAGES C CXX)

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Detect build environment
if(CMAKE_CROSSCOMPILING)
    message(STATUS "Cross-compiling for ${CMAKE_SYSTEM_PROCESSOR}")
else()
    message(STATUS "Native build for ${CMAKE_HOST_SYSTEM_PROCESSOR}")
endif()

# Rockchip SDK paths - Updated for RK3566 with local SDK
set(RK_SDK_BASE "${CMAKE_CURRENT_SOURCE_DIR}/sdk")
set(RK_ROCKIT_PATH "${RK_SDK_BASE}/vendor/rockit")
set(RK_MPI_INCLUDE_PATH "${RK_ROCKIT_PATH}/mpi/sdk/include")
set(RK_TGI_INCLUDE_PATH "${RK_ROCKIT_PATH}/tgi/sdk/include")
set(RK_MPP_INCLUDE_PATH "${RK_SDK_BASE}/mpp/include")

# Determine architecture-specific library path
# For RK3566 (ARM64), always use lib64, but handle both cross-compilation and native ARM64
if(CMAKE_SYSTEM_PROCESSOR MATCHES "aarch64|arm64" OR CMAKE_CROSSCOMPILING)
    set(RK_LIB_PATH "${RK_ROCKIT_PATH}/lib/lib64")
    message(STATUS "Using ARM64 libraries from ${RK_LIB_PATH}")
else()
    # Fallback for other architectures (though this project targets ARM64)
    set(RK_LIB_PATH "${RK_ROCKIT_PATH}/lib/lib32")
    message(WARNING "Using lib32 - this may not work correctly for RK3566")
endif()

set(RK_MPP_LIB_PATH "${RK_SDK_BASE}/mpp/lib")

# LVGL removed for headless operation

# Get source files, excluding CMake generated files and sysroot documentation
file(GLOB_RECURSE sources CONFIGURE_DEPENDS "*.c")
list(FILTER sources EXCLUDE REGEX "CMakeFiles.*CompilerId.*\\.c$")
# Exclude main.c from library sources (it's used for the binary target)
list(FILTER sources EXCLUDE REGEX "main\\.c$")
# Exclude sysroot documentation and example files
list(FILTER sources EXCLUDE REGEX "sysroot/.*\\.c$")

add_library(jknative STATIC ${sources} ${CMAKE_CURRENT_SOURCE_DIR}/ctrl.h)

# Include directories
target_include_directories(jknative PRIVATE
    ${RK_MPI_INCLUDE_PATH}
    ${RK_TGI_INCLUDE_PATH}
    ${RK_MPP_INCLUDE_PATH}
    ${CMAKE_CURRENT_SOURCE_DIR}
    ${CMAKE_CURRENT_BINARY_DIR}/deps/libgpiod/include
)

# Set library search path
target_link_directories(jknative PRIVATE 
    ${RK_LIB_PATH}
    ${RK_MPP_LIB_PATH}
)
# target_link_directories(jknative PRIVATE ${CMAKE_CURRENT_BINARY_DIR}/deps/libgpiod/lib)

target_link_libraries(jknative PRIVATE
    pthread
    rockit
    rockchip_mpp
    m
    # libgpiod
)

# Binary target using main.c as entry point
option(BUILD_BINARY "Build the binary executable" ON)

# Only build binary if we have compatible libraries or are cross-compiling
if(BUILD_BINARY AND (CMAKE_CROSSCOMPILING OR CMAKE_SYSTEM_PROCESSOR MATCHES "aarch64|arm64"))
    add_executable(jknative-bin ${CMAKE_CURRENT_SOURCE_DIR}/main.c)

    # Set library search path for the binary as well
    target_link_directories(jknative-bin PRIVATE 
        ${RK_LIB_PATH}
        ${RK_MPP_LIB_PATH}
    )

    # Link the binary to the library
    target_link_libraries(jknative-bin PRIVATE
        jknative
        pthread
        rockit
        rockchip_mpp
        m
    )

    install(TARGETS jknative-bin DESTINATION bin)
    message(STATUS "Binary target enabled")
else()
    message(STATUS "Binary target disabled - incompatible architecture or BUILD_BINARY=OFF")
endif()
install(TARGETS jknative DESTINATION lib)