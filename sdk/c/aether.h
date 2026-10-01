#ifndef AETHER_H
#define AETHER_H
#include <stddef.h>
/* POSIX blocking client. Returns fd or -1 with errno. Timeout applies to
 * send/receive, not DNS resolution or connect. Caller owns the descriptor. */
int aether_connect(const char *host, const char *port, int timeout_seconds);
/* Sends one v1 data frame; 0 success, -1 error. No acknowledgements/retries. */
int aether_send(int fd, const void *payload, size_t length);
int aether_close(int fd);
#endif
