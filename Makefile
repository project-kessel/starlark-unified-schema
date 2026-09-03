.PHONY: build-interpreter build-interpreter-debug build-schema build-shipped-schema package-release clean test

# Features-only delivery allowlist. HBI/RBAC stay in schema/ for full
# build-schema / build-test coverage, but are not shipped to consumers yet —
# those namespaces remain hand-authored downstream. Drop this list (full walk)
# once those services onboard.
SHIPPED_SCHEMA_FILES ?= \
	service/reporters/features/service.star \
	billing_account/reporters/features/billing_account.star \
	workspace/reporters/features/workspace.star

# KSIL files included in ksl.tar.gz. Do not ship hbi.json / rbac.json until
# those namespaces migrate off hand-authored .ksl in rbac-config.
KSL_RELEASE_FILES ?= features.json

JSONSCHEMA_OUTPUT_DIR ?= output/jsonschema
KSL_OUTPUT_DIR ?= output/ksl

build-interpreter:
	go build -C ./interpreter/ -o ../bin/interpreter ./cmd/interpreter

build-interpreter-debug:
	go build -C ./interpreter/ -o ../bin/interpreter -gcflags="all=-N -l" ./cmd/interpreter

test:
	go test -C ./interpreter/ -count=1 ./...

build-schema: build-interpreter
	dotenv -f .env run ./bin/interpreter

# Build only the artifacts that are delivered to consumers.
build-shipped-schema: build-interpreter
	rm -rf "$(JSONSCHEMA_OUTPUT_DIR)" "$(KSL_OUTPUT_DIR)"
	JSONSCHEMA_OUTPUT_DIR="$(JSONSCHEMA_OUTPUT_DIR)" KSL_OUTPUT_DIR="$(KSL_OUTPUT_DIR)" \
		./bin/interpreter $(SHIPPED_SCHEMA_FILES)

# Package release tarballs in the repo root (gitignored).
package-release: build-shipped-schema
	@test -s "$(KSL_OUTPUT_DIR)/features.json" || { \
		echo "error: missing or empty $(KSL_OUTPUT_DIR)/features.json" >&2; \
		exit 1; \
	}
	tar czf jsonschema.tar.gz -C "$(JSONSCHEMA_OUTPUT_DIR)" .
	tar czf ksl.tar.gz -C "$(KSL_OUTPUT_DIR)" $(KSL_RELEASE_FILES)

clean:
	rm -rf bin/
	rm -rf output/
	rm -f jsonschema.tar.gz ksl.tar.gz
