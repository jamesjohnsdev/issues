MANDIR ?= /usr/local/share/man/man1

.PHONY: man install-man

man:
	go run ./tools/gen-man man

install-man: man
	install -Dm644 man/issues*.1 -t $(MANDIR)
