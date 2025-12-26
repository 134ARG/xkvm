cmake_minimum_required(VERSION 3.14)
include(FetchContent)
include(ExternalProject)

project(jknative LANGUAGES C CXX)

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Rockchip SDK paths - Updated for RK3566 with local SDK
set(RK_SDK_BASE "${CMAKE_CURRENT_SOURCE_DIR}/sdk")
set(RK_ROCKIT_PATH "${RK_SDK_BASE}/vendor/rockit")
set(RK_MPI_INCLUDE_PATH "${RK_ROCKIT_PATH}/mpi/sdk/include")
set(RK_TGI_INCLUDE_PATH "${RK_ROCKIT_PATH}/tgi/sdk/include")
set(RK_MPP_INCLUDE_PATH "${RK_SDK_BASE}/mpp/include")

# Determine architecture-specific library path
# For RK3566 (ARM64), always use lib64
set(RK_LIB_PATH "${RK_ROCKIT_PATH}/lib/lib64")

set(RK_MPP_LIB_PATH "${RK_SDK_BASE}/mpp/lib")

# LVGL removed for headless operation

# Get source files, excluding CMake generated files
file(GLOB_RECURSE sources CONFIGURE_DEPENDS "*.c")
list(FILTER sources EXCLUDE REGEX "CMakeFiles.*CompilerId.*\\.c$")
# Exclude main.c from library sources (it's used for the binary target)
list(FILTER sources EXCLUDE REGEX "main\\.c$")

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

# Binary target using main.c as entry point (optional for cross-compilation)
option(BUILD_BINARY "Build the binary executable" ON)
if(BUILD_BINARY)
    add_executable(jknative-bin ${CMAKE_CURRENT_SOURCE_DIR}/main.c)

    # Set library search path for the binary as well
    target_link_directories(jknative-bin PRIVATE 
        ${RK_LIB_PATH}
        ${RK_MPP_LIB_PATH}
    )

    # Link the binary to the library (if needed in the future)
    target_link_libraries(jknative-bin PRIVATE
        jknative
        pthread
        rockit
        rockchip_mpp
        m
    )

    install(TARGETS jknative-bin DESTINATION bin)
endif()
install(TARGETS jknative DESTINATION lib)