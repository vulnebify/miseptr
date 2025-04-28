APP_NAME = miseptr
CMD_PATH = ./cmd/miseptr
BUILD_DIR = ./bin
VERSION ?= $(shell git describe --tags --always --dirty)

KUBERNETES_VERSION = 1.27
ENVTEST_BIN_ROOT = /tmp/testbin
KUBEBUILDER_ASSETS_PATH = $(shell setup-envtest use $(KUBERNETES_VERSION) --bin-dir=$(ENVTEST_BIN_ROOT) | grep 'Path:' | awk '{print $$2}')

.PHONY: all build clean test setup-envtest fetch-envtest-binaries

all: build

build:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags "-X=github.com/vulnebify/miseptr/internal.Version=$(VERSION)" \
	-o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)

clean:
	rm -rf $(BUILD_DIR)

setup-envtest:
	go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
	@echo "✅ setup-envtest installed"

fetch-envtest-binaries:
	@echo "🌐 Downloading Kubernetes binaries for envtest..."
	mkdir -p $(ENVTEST_BIN_ROOT)
	@setup-envtest use $(KUBERNETES_VERSION) --bin-dir=$(ENVTEST_BIN_ROOT)
	@echo "✅ Binaries fetched into $(KUBEBUILDER_ASSETS_PATH)"

test: fetch-envtest-binaries
	@echo "🚀 Running tests..."
	@echo "Using KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS_PATH)"
	export KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS_PATH) && \
	go test ./internal/controller -v
