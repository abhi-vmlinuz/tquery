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
	@mkdir -p $(DESTDIR)$(BINDIR)
	@install -m 755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY) || (echo "Permission denied writing to $(DESTDIR)$(BINDIR). Run with sudo or use 'make install-user'" && exit 1)
	@ln -sf $(BINARY) $(DESTDIR)$(BINDIR)/$(ALIAS)
	@mkdir -p $(DESTDIR)$(MANDIR)
	@install -m 644 man/tq.1 $(DESTDIR)$(MANDIR)/tq.1
	@ln -sf tq.1 $(DESTDIR)$(MANDIR)/tquery.1
	@# Install Bash completions to standard system locations
	@mkdir -p $(DESTDIR)$(BASHCOMPDIR)
	@install -m 644 completions/tq.bash $(DESTDIR)$(BASHCOMPDIR)/tq
	@ln -sf tq $(DESTDIR)$(BASHCOMPDIR)/tquery
	@mkdir -p $(DESTDIR)/etc/bash_completion.d 2>/dev/null || true
	@install -m 644 completions/tq.bash $(DESTDIR)/etc/bash_completion.d/tq 2>/dev/null || true
	@ln -sf tq $(DESTDIR)/etc/bash_completion.d/tquery 2>/dev/null || true
	@mkdir -p $(DESTDIR)/usr/share/bash-completion/completions 2>/dev/null || true
	@install -m 644 completions/tq.bash $(DESTDIR)/usr/share/bash-completion/completions/tq 2>/dev/null || true
	@ln -sf tq $(DESTDIR)/usr/share/bash-completion/completions/tquery 2>/dev/null || true
	@# Install Fish completions
	@mkdir -p $(DESTDIR)$(FISHCOMPDIR)
	@install -m 644 completions/tq.fish $(DESTDIR)$(FISHCOMPDIR)/tq.fish
	@ln -sf tq.fish $(DESTDIR)$(FISHCOMPDIR)/tquery.fish
	@mkdir -p $(DESTDIR)/usr/share/fish/vendor_completions.d 2>/dev/null || true
	@install -m 644 completions/tq.fish $(DESTDIR)/usr/share/fish/vendor_completions.d/tq.fish 2>/dev/null || true
	@ln -sf tq.fish $(DESTDIR)/usr/share/fish/vendor_completions.d/tquery.fish 2>/dev/null || true
	@# Install Zsh completions
	@mkdir -p $(DESTDIR)$(ZSHCOMPDIR)
	@install -m 644 completions/tq.zsh $(DESTDIR)$(ZSHCOMPDIR)/_tq
	@ln -sf _tq $(DESTDIR)$(ZSHCOMPDIR)/_tquery
	@mkdir -p $(DESTDIR)/usr/share/zsh/site-functions 2>/dev/null || true
	@install -m 644 completions/tq.zsh $(DESTDIR)/usr/share/zsh/site-functions/_tq 2>/dev/null || true
	@ln -sf _tq $(DESTDIR)/usr/share/zsh/site-functions/_tquery 2>/dev/null || true
	@echo "[ok] System installation complete."

install-user: build
	@mkdir -p $(HOME)/.local/bin
	@install -m 755 $(BINARY) $(HOME)/.local/bin/$(BINARY)
	@ln -sf $(BINARY) $(HOME)/.local/bin/$(ALIAS)
	@mkdir -p $(HOME)/.local/share/man/man1
	@install -m 644 man/tq.1 $(HOME)/.local/share/man/man1/tq.1
	@ln -sf tq.1 $(HOME)/.local/share/man/man1/tquery.1
	@# User Bash completion
	@mkdir -p $(HOME)/.local/share/bash-completion/completions
	@install -m 644 completions/tq.bash $(HOME)/.local/share/bash-completion/completions/tq
	@ln -sf tq $(HOME)/.local/share/bash-completion/completions/tquery
	@mkdir -p $(HOME)/.bash_completion.d
	@install -m 644 completions/tq.bash $(HOME)/.bash_completion.d/tq
	@ln -sf tq $(HOME)/.bash_completion.d/tquery
	@# User Fish completion
	@mkdir -p $(HOME)/.config/fish/completions
	@install -m 644 completions/tq.fish $(HOME)/.config/fish/completions/tq.fish
	@ln -sf tq.fish $(HOME)/.config/fish/completions/tquery.fish
	@# User Zsh completion
	@mkdir -p $(HOME)/.local/share/zsh/site-functions
	@install -m 644 completions/tq.zsh $(HOME)/.local/share/zsh/site-functions/_tq
	@ln -sf _tq $(HOME)/.local/share/zsh/site-functions/_tquery
	@mkdir -p $(HOME)/.zsh/completion
	@install -m 644 completions/tq.zsh $(HOME)/.zsh/completion/_tq
	@ln -sf _tq $(HOME)/.zsh/completion/_tquery
	@echo "[ok] User installation complete."

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY) $(DESTDIR)$(BINDIR)/$(ALIAS)
	rm -f $(DESTDIR)$(MANDIR)/tq.1 $(DESTDIR)$(MANDIR)/tquery.1
	rm -f $(DESTDIR)$(BASHCOMPDIR)/tq $(DESTDIR)$(BASHCOMPDIR)/tquery
	rm -f $(DESTDIR)/etc/bash_completion.d/tq $(DESTDIR)/etc/bash_completion.d/tquery
	rm -f $(DESTDIR)/usr/share/bash-completion/completions/tq $(DESTDIR)/usr/share/bash-completion/completions/tquery
	rm -f $(DESTDIR)$(FISHCOMPDIR)/tq.fish $(DESTDIR)$(FISHCOMPDIR)/tquery.fish
	rm -f $(DESTDIR)/usr/share/fish/vendor_completions.d/tq.fish $(DESTDIR)/usr/share/fish/vendor_completions.d/tquery.fish
	rm -f $(DESTDIR)$(ZSHCOMPDIR)/_tq $(DESTDIR)$(ZSHCOMPDIR)/_tquery
	rm -f $(DESTDIR)/usr/share/zsh/site-functions/_tq $(DESTDIR)/usr/share/zsh/site-functions/_tquery
	rm -f $(HOME)/.local/bin/$(BINARY) $(HOME)/.local/bin/$(ALIAS)
	rm -f $(HOME)/.local/share/man/man1/tq.1 $(HOME)/.local/share/man/man1/tquery.1
	rm -f $(HOME)/.local/share/bash-completion/completions/tq $(HOME)/.local/share/bash-completion/completions/tquery
	rm -f $(HOME)/.bash_completion.d/tq $(HOME)/.bash_completion.d/tquery
	rm -f $(HOME)/.config/fish/completions/tq.fish $(HOME)/.config/fish/completions/tquery.fish
	rm -f $(HOME)/.local/share/zsh/site-functions/_tq $(HOME)/.local/share/zsh/site-functions/_tquery
	rm -f $(HOME)/.zsh/completion/_tq $(HOME)/.zsh/completion/_tquery
