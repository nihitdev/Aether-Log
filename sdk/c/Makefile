CC ?= cc
CFLAGS ?= -O2 -std=c11 -Wall -Wextra -Werror
.PHONY: all clean
all: example
example: example.c aether.c aether.h
	$(CC) $(CFLAGS) example.c aether.c -o example
clean:
	rm -f example
