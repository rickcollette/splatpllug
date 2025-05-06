# -------------------------------------------------------------------
# examples: build loader into bin/loader and echo plugin into plugins/echo
BIN_DIR     := bin
PLUGIN_DIR  := plugins
EXAMPLES    := echo loader

.PHONY: examples
examples: $(BIN_DIR)/loader $(PLUGIN_DIR)/echo

# ensure output dirs exist
$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(PLUGIN_DIR):
	mkdir -p $(PLUGIN_DIR)

# build the loader CLI
$(BIN_DIR)/loader: examples/loader/main.go | $(BIN_DIR)
	go build -o $@ $<

# build the echo plugin
$(PLUGIN_DIR)/echo: examples/echo/main.go | $(PLUGIN_DIR)
	go build -o $@ $<

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) $(PLUGIN_DIR)
