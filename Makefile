# -------------------------------------------------------------------
# Build loader CLI and all example plugins:
BIN_DIR    := bin
PLUGIN_DIR := plugins

# List of example plugins (must match folders under examples/)
PLUGIN_EXAMPLES := echo math reverse rolodex

.PHONY: examples
examples: $(BIN_DIR)/loader $(PLUGIN_EXAMPLES:%=$(PLUGIN_DIR)/%)

# ensure output dirs exist
$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(PLUGIN_DIR):
	mkdir -p $(PLUGIN_DIR)

# build the loader CLI
$(BIN_DIR)/loader: examples/loader/main.go | $(BIN_DIR)
	go build -o $@ $<

# build each plugin into plugins/<name>
$(PLUGIN_DIR)/%: examples/%/main.go | $(PLUGIN_DIR)
	go build -o $@ $<

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) $(PLUGIN_DIR)
