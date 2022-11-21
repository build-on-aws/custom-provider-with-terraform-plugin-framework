HOSTNAME=aws.amazon.com
NAMESPACE=terraform
NAME=buildonaws
VERSION=1.0

OS_NAME:=$(shell uname -s | tr ‘[:upper:]’ ‘[:lower:]’)
HW_CLASS:=$(shell uname -m)
OS_ARCH=${OS_NAME}_${HW_CLASS}

BINARY=terraform-provider-${NAME}
PLUGIN_DIR=${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

default: install

build: clean
	@echo "▶️▶️▶️ Building the Terraform binary file"
	go build -gcflags="all=-N -l" -o ${BINARY}

install: build
	@echo "▶️▶️▶️ Adding the binary to the plugin directory"
	mkdir -p ~/.terraform.d/plugins/${PLUGIN_DIR}
	mv ${BINARY} ~/.terraform.d/plugins/${PLUGIN_DIR}
	@echo "▶️▶️▶️ Build executed successfully"

clean:
	@echo "▶️▶️▶️ Removing the Terraform plugin"
	rm -rf ~/.terraform.d/plugins/${PLUGIN_DIR}
	rm -rf examples/.terraform* || true

test:
	TF_ACC=1 go test -count=1 -parallel=4 -timeout 5m -v ./${NAME}

generate:
	go generate ./...
