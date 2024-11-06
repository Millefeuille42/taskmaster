ifeq ($(origin .RECIPEPREFIX), undefined)
    $(error This Make does not support .RECIPEPREFIX.
    Please use GNU Make 4.0 or later)
endif
.RECIPEPREFIX = >

.POSIX:
.SUFFIXES: .ha
HARE=hare
HAREFLAGS=

DESTDIR=
PREFIX=/usr/local
BINDIR=$(PREFIX)/bin

NAME=hareForceOne

SRCS=$(shell find ./cmd/${NAME} -name '*.ha')

all: $(NAME)

$(NAME): $(SRCS)
> cd cmd/$@/ && $(HARE) build $(HAREFLAGS) -o $(PWD)/$@ .

check:
> $(HARE) test $(HAREFLAGS)

clean:
> rm -f $(NAME)

install:
> install -Dm755 $(NAME) $(DESTDIR)$(BINDIR)/$(NAME)

uninstall:
> rm -f $(DESTDIR)$(BINDIR)/$(NAME)

run: all
> ./$(NAME)

.PHONY: all check clean install uninstall run