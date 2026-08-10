PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
MANDIR ?= $(PREFIX)/share/man/man1
BASHCOMPDIR ?= $(PREFIX)/share/bash-completion/completions
FISHCOMPDIR ?= $(PREFIX)/share/fish/vendor_completions.d
ZSHCOMPDIR ?= $(PREFIX)/share/zsh/site-functions

BINARY = tq
ALIAS = tquery
SRC = $(shell find . -name "*.go")

.PHONY: all build test clean install install-user uninstall

all: test build

build: $(BINARY)

$(BINARY): $(SRC)
	go build -ldflags="-s -w" -o $(BINARY) main.go
	ln -sf $(BINARY) $(ALIAS)

test:
	go test -v ./...

clean:
	rm -f $(BINARY) $(ALIAS)

install: build
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	ln -sf $(BINARY) $(DESTDIR)$(BINDIR)/$(ALIAS)
	install -d $(DESTDIR)$(MANDIR)
	install -m 644 man/tq.1 $(DESTDIR)$(MANDIR)/tq.1
	ln -sf tq.1 $(DESTDIR)$(MANDIR)/tquery.1
	install -d $(DESTDIR)$(BASHCOMPDIR)
	install -m 644 completions/tq.bash $(DESTDIR)$(BASHCOMPDIR)/tq
	ln -sf tq $(DESTDIR)$(BASHCOMPDIR)/tquery
	install -d $(DESTDIR)$(FISHCOMPDIR)
	install -m 644 completions/tq.fish $(DESTDIR)$(FISHCOMPDIR)/tq.fish
	ln -sf tq.fish $(DESTDIR)$(FISHCOMPDIR)/tquery.fish
	install -d $(DESTDIR)$(ZSHCOMPDIR)
	install -m 644 completions/tq.zsh $(DESTDIR)$(ZSHCOMPDIR)/_tq
	ln -sf _tq $(DESTDIR)$(ZSHCOMPDIR)/_tquery

install-user: build
	install -d $(HOME)/.local/bin
	install -m 755 $(BINARY) $(HOME)/.local/bin/$(BINARY)
	ln -sf $(BINARY) $(HOME)/.local/bin/$(ALIAS)
	install -d $(HOME)/.local/share/man/man1
	install -m 644 man/tq.1 $(HOME)/.local/share/man/man1/tq.1
	ln -sf tq.1 $(HOME)/.local/share/man/man1/tquery.1
	install -d $(HOME)/.config/fish/completions
	install -m 644 completions/tq.fish $(HOME)/.config/fish/completions/tq.fish
	ln -sf tq.fish $(HOME)/.config/fish/completions/tquery.fish
	install -d $(HOME)/.local/share/bash-completion/completions
	install -m 644 completions/tq.bash $(HOME)/.local/share/bash-completion/completions/tq
	ln -sf tq $(HOME)/.local/share/bash-completion/completions/tquery

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY) $(DESTDIR)$(BINDIR)/$(ALIAS)
	rm -f $(DESTDIR)$(MANDIR)/tq.1 $(DESTDIR)$(MANDIR)/tquery.1
	rm -f $(DESTDIR)$(BASHCOMPDIR)/tq $(DESTDIR)$(BASHCOMPDIR)/tquery
	rm -f $(DESTDIR)$(FISHCOMPDIR)/tq.fish $(DESTDIR)$(FISHCOMPDIR)/tquery.fish
	rm -f $(DESTDIR)$(ZSHCOMPDIR)/_tq $(DESTDIR)$(ZSHCOMPDIR)/_tquery
