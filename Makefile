BRANCH    := $(shell git rev-parse --abbrev-ref HEAD)
BUILDDATE := $(shell date -u +%FT%T%z)
BUILDTS   := $(shell date -u +%s)
REVISION  := $(shell git rev-parse HEAD)
VERSION := 0.1.2
VERSION_DEV := $(VERSION)-dev$(shell date -u +%Y%m%d%H%M)

PROMETHEUS_TAG := github.com/prometheus/common/version
KVM_PKG_NAME := github.com/134ARG/xkvm

SKIP_NATIVE_IF_EXISTS ?= 0
SKIP_UI_BUILD ?= 0
ENABLE_SYNC_TRACE ?= 0

CMAKE_BUILD_TYPE ?= Release

GO_BUILD_ARGS := -tags netgo,timetzdata,nomsgpack
ifeq ($(ENABLE_SYNC_TRACE), 1)
	GO_BUILD_ARGS := $(GO_BUILD_ARGS),synctrace
endif

GO_RELEASE_BUILD_ARGS := -trimpath $(GO_BUILD_ARGS)
GO_LDFLAGS := \
  -s -w \
  -extldflags '-Wl,-rpath,\$$ORIGIN/lib' \
  -X $(PROMETHEUS_TAG).Branch=$(BRANCH) \
  -X $(PROMETHEUS_TAG).BuildDate=$(BUILDDATE) \
  -X $(PROMETHEUS_TAG).Revision=$(REVISION) \
  -X $(KVM_PKG_NAME).builtTimestamp=$(BUILDTS)

# ARM64 cross-compilation configuration
# ARM64_SYSROOT must be set for cross-compilation

GO_ARGS := GOOS=linux GOARCH=arm64 CGO_ENABLED=1 ARCHFLAGS="-arch arm64" \
	CC=aarch64-linux-gnu-gcc \
	CXX=aarch64-linux-gnu-g++ \
	ARM64_SYSROOT=$(ARM64_SYSROOT) \
	CGO_CFLAGS="--sysroot=$(ARM64_SYSROOT) -I$(ARM64_SYSROOT)/usr/include/aarch64-linux-gnu" \
	CGO_LDFLAGS="--sysroot=$(ARM64_SYSROOT) -B$(ARM64_SYSROOT)/lib64 -L$(ARM64_SYSROOT)/lib64 -L$(ARM64_SYSROOT)/usr/lib/aarch64-linux-gnu -L$(shell pwd)/internal/native/cgo/sdk/vendor/rockit/lib/lib64 -L$(shell pwd)/internal/native/cgo/sdk/mpp/lib"

GO_CMD := $(GO_ARGS) go

BIN_DIR := $(shell pwd)/bin

TEST_DIRS := $(shell find . -name "*_test.go" -type f -exec dirname {} \; | sort -u)

test:
	go test ./...

test_e2e:
	@read -p "Device IP: " device_ip; \
	cd ui && npm install && npx playwright install --with-deps chromium && \
	NODE_NO_WARNINGS=1 XKVM_URL="http://$$device_ip" npm run test:e2e

lint:
	go vet ./...

check: lint test

build_native:
	@if [ "$(SKIP_NATIVE_IF_EXISTS)" = "1" ] && [ -f "internal/native/cgo/lib/libjknative.a" ]; then \
		echo "libjknative.a already exists, skipping native build..."; \
	else \
		echo "Building native..."; \
		CMAKE_BUILD_TYPE=$(CMAKE_BUILD_TYPE) ./scripts/build_cgo.sh; \
	fi

build_dev:
	@if [ -z "$(ARM64_SYSROOT)" ]; then \
		echo "❌ Error: ARM64 cross-compilation requires ARM64_SYSROOT environment variable"; \
		echo "💡 Set ARM64_SYSROOT to your ARM64 sysroot path:"; \
		echo "   export ARM64_SYSROOT=\"/path/to/your/arm64-sysroot\""; \
		echo "   make build_dev"; \
		echo ""; \
		echo "💡 Or create a sysroot with:"; \
		echo "   ./scripts/setup_arm64_sysroot.sh"; \
		exit 1; \
	elif [ ! -d "$(ARM64_SYSROOT)" ]; then \
		echo "❌ Error: ARM64_SYSROOT path does not exist: $(ARM64_SYSROOT)"; \
		echo "💡 Create the sysroot with:"; \
		echo "   ARM64_SYSROOT=\"$(ARM64_SYSROOT)\" ./scripts/setup_arm64_sysroot.sh"; \
		exit 1; \
	fi
	$(MAKE) _build_dev_inner VERSION_DEV=$(VERSION_DEV)
		exit 1; \
	elif [ ! -d "$(ARM64_SYSROOT)" ]; then \
		echo "❌ Error: ARM64_SYSROOT path does not exist: $(ARM64_SYSROOT)"; \
		echo "💡 Create the sysroot with:"; \
		echo "   ARM64_SYSROOT=\"$(ARM64_SYSROOT)\" ./scripts/setup_arm64_sysroot.sh"; \
		exit 1; \
	fi
	$(MAKE) _build_dev_inner VERSION_DEV=$(VERSION_DEV)

_build_dev_inner: build_native
	@echo "Building... $(VERSION_DEV)"
	$(GO_CMD) build \
		-ldflags="$(GO_LDFLAGS) -X $(KVM_PKG_NAME).builtAppVersion=$(VERSION_DEV)" \
		$(GO_RELEASE_BUILD_ARGS) \
		-o $(BIN_DIR)/xkvm_app -v cmd/main.go

build_test2json:
	$(GO_CMD) build -o $(BIN_DIR)/test2json cmd/test2json

build_gotestsum:
	@echo "Building gotestsum..."
	$(GO_CMD) install gotest.tools/gotestsum@latest
	cp $(shell $(GO_CMD) env GOPATH)/bin/linux_arm/gotestsum $(BIN_DIR)/gotestsum

build_dev_test: build_test2json build_gotestsum
# collect all directories that contain tests
	@echo "Building tests for devices ..."
	@rm -rf $(BIN_DIR)/tests && mkdir -p $(BIN_DIR)/tests

	@cat resource/dev_test.sh > $(BIN_DIR)/tests/run_all_tests
	@for test in $(TEST_DIRS); do \
		test_pkg_name=$$(echo $$test | sed 's/^.\///g'); \
		test_pkg_full_name=$(KVM_PKG_NAME)/$$(echo $$test | sed 's/^.\///g'); \
		test_filename=$$(echo $$test_pkg_name | sed 's/\//__/g')_test; \
		$(GO_CMD) test -v \
			-ldflags="$(GO_LDFLAGS) -X $(KVM_PKG_NAME).builtAppVersion=$(VERSION_DEV)" \
			$(GO_BUILD_ARGS) \
			-c -o $(BIN_DIR)/tests/$$test_filename $$test; \
		echo "runTest ./$$test_filename $$test_pkg_full_name" >> $(BIN_DIR)/tests/run_all_tests; \
	done; \
	chmod +x $(BIN_DIR)/tests/run_all_tests; \
	cp $(BIN_DIR)/test2json $(BIN_DIR)/tests/ && chmod +x $(BIN_DIR)/tests/test2json; \
	cp $(BIN_DIR)/gotestsum $(BIN_DIR)/tests/ && chmod +x $(BIN_DIR)/tests/gotestsum; \
	tar czfv device-tests.tar.gz -C $(BIN_DIR)/tests .

frontend:
	@if [ "$(SKIP_UI_BUILD)" = "1" ] && [ -f "static/index.html" ]; then \
		echo "Skipping frontend build..."; \
	else \
		cd ui && npm ci && npm run build:device && \
		find ../static/ -type f \
			\( -name '*.js' \
			-o -name '*.css' \
			-o -name '*.html' \
			-o -name '*.ico' \
			-o -name '*.png' \
			-o -name '*.jpg' \
			-o -name '*.jpeg' \
			-o -name '*.gif' \
			-o -name '*.svg' \
			-o -name '*.webp' \
			-o -name '*.woff2' \
			\) -exec sh -c 'gzip -9 -kfv {}' \; ;\
	fi

build_release:
	@if [ -z "$(ARM64_SYSROOT)" ]; then \
		echo "❌ Error: ARM64 cross-compilation requires ARM64_SYSROOT environment variable"; \
		echo "💡 Set ARM64_SYSROOT to your ARM64 sysroot path:"; \
		echo "   export ARM64_SYSROOT=\"/path/to/your/arm64-sysroot\""; \
		echo "   make build_release"; \
		echo ""; \
		echo "💡 Or create a sysroot with:"; \
		echo "   ./scripts/setup_arm64_sysroot.sh"; \
		exit 1; \
	elif [ ! -d "$(ARM64_SYSROOT)" ]; then \
		echo "❌ Error: ARM64_SYSROOT path does not exist: $(ARM64_SYSROOT)"; \
		echo "💡 Create the sysroot with:"; \
		echo "   ARM64_SYSROOT=\"$(ARM64_SYSROOT)\" ./scripts/setup_arm64_sysroot.sh"; \
		exit 1; \
	fi
	$(MAKE) _build_release_inner VERSION=$(VERSION)

_build_release_inner: build_native
	@echo "Building release..."
	$(GO_CMD) build \
		-ldflags="$(GO_LDFLAGS) -X $(KVM_PKG_NAME).builtAppVersion=$(VERSION)" \
		$(GO_RELEASE_BUILD_ARGS) \
		-o bin/xkvm_app cmd/main.go
	@echo "Creating self-extracting installer..."
	@./scripts/create_self_extract.sh

build_packages: build_release
	@echo "Building DEB and RPM packages..."
	@./scripts/build_packages.sh
