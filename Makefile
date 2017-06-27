
INST_PATH:=$(GOPATH)/bin/sunfish4-ga
BIN:=bin/sunfish4-ga
PKG:=github.com/sunfish-shogi/sunfish4-ga

.PHONY: help
help:
	@echo "USAGE:"
	@echo "  make build"
	@echo "  make install"
	@echo "  make clean"

.PHONY: install
install: $(INST_PATH)

.PHONY: build
build: $(BIN)

.PHONY: vet
vet:
	go vet $(PKG)/...

$(INST_PATH): $(BIN)
	cp $(BIN) $(INST_PATH)

.PHONY: $(BIN)
$(BIN):
	go build -o $(BIN) $(PKG)

.PHONY: $(CLEAN)
$(CLEAN):
	$(RM) $(BIN)
