GO ?= go

DIST_DIR := dist
WASM_OUT := $(DIST_DIR)/renderer.wasm
WASM_EXEC_OUT := $(DIST_DIR)/wasm_exec.js

.PHONY: help clean server wasm dist

help:
	@echo "Targets:"
	@echo "  make server  - run the HTTP server (./server)"
	@echo "  make wasm    - build wasm + copy wasm_exec.js into ./dist"
	@echo "  make clean   - remove ./dist"

clean:
	rm -rf $(DIST_DIR)

server:
	$(GO) run ./server

dist:
	mkdir -p $(DIST_DIR)

wasm: dist
	GOOS=js GOARCH=wasm $(GO) build -o $(WASM_OUT) ./wasm
	cp "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" $(WASM_EXEC_OUT)
