.POSIX:
.SUFFIXES: .ha
HARE=hare
HAREFLAGS=

DESTDIR=
PREFIX=/usr/local
BINDIR=$(PREFIX)/bin

NAME=harels

SRCS=cmd/$(NAME)/main.ha

all: $(NAME)

$(NAME): $(SRCS)
	$(HARE) build $(HAREFLAGS) -o $@ cmd/$@/

check:
	$(HARE) test $(HAREFLAGS)

clean:
	rm -f $(NAME)

install:
	install -Dm755 $(NAME) $(DESTDIR)$(BINDIR)/$(NAME)

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(NAME)

.PHONY: all check clean install uninstall