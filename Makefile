PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
MANDIR ?= $(PREFIX)/share/man/man1
BASHCOMPDIR ?= $(PREFIX)/share/bash-completion/completions
FISHCOMPDIR ?= $(PREFIX)/share/fish/vendor_completions.d
ZSHCOMPDIR ?= $(PREFIX)/share/zsh/site-functions

BINARY = tquery
SRC = $(shell find . -name "*.go")

.PHONY: all build test clean install install-user uninstall

all: test build

build: $(BINARY)

$(BINARY): $(SRC)
	go build -ldflags="-s -w" -o $(BINARY) main.go

test:
	go test -v ./...

clean:
	rm -f $(BINARY)

install: build
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	install -d $(DESTDIR)$(MANDIR)
	install -m 644 man/tquery.1 $(DESTDIR)$(MANDIR)/tquery.1
	install -d $(DESTDIR)$(BASHCOMPDIR)
	install -m 644 completions/tquery.bash $(DESTDIR)$(BASHCOMPDIR)/tquery
	install -d $(DESTDIR)$(FISHCOMPDIR)
	install -m 644 completions/tquery.fish $(DESTDIR)$(FISHCOMPDIR)/tquery.fish
	install -d $(DESTDIR)$(ZSHCOMPDIR)
	install -m 644 completions/tquery.zsh $(DESTDIR)$(ZSHCOMPDIR)/_tquery

install-user: build
	install -d $(HOME)/.local/bin
	install -m 755 $(BINARY) $(HOME)/.local/bin/$(BINARY)
	install -d $(HOME)/.local/share/man/man1
	install -m 644 man/tquery.1 $(HOME)/.local/share/man/man1/tquery.1
	install -d $(HOME)/.config/fish/completions
	install -m 644 completions/tquery.fish $(HOME)/.config/fish/completions/tquery.fish
	install -d $(HOME)/.local/share/bash-completion/completions
	install -m 644 completions/tquery.bash $(HOME)/.local/share/bash-completion/completions/tquery

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)
	rm -f $(DESTDIR)$(MANDIR)/tquery.1
	rm -f $(DESTDIR)$(BASHCOMPDIR)/tquery
	rm -f $(DESTDIR)$(FISHCOMPDIR)/tquery.fish
	rm -f $(DESTDIR)$(ZSHCOMPDIR)/_tquery
