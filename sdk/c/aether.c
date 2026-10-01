#define _POSIX_C_SOURCE 200112L
#include "aether.h"
#include <errno.h>
#include <stdint.h>
#include <netdb.h>
#include <sys/socket.h>
#include <sys/time.h>
#include <unistd.h>
#include <string.h>
static int send_all(int fd, const unsigned char *p, size_t n) {
 while(n) {
#ifdef MSG_NOSIGNAL
  ssize_t sent=send(fd,p,n,MSG_NOSIGNAL);
#else
  ssize_t sent=send(fd,p,n,0);
#endif
  if(sent<0) { if(errno==EINTR) continue; return -1; }
  if(sent==0) { errno=EPIPE; return -1; }
  p+=sent; n-=(size_t)sent;
 }
 return 0;
}
int aether_connect(const char *host,const char *port,int seconds) {
 if(!host || !port || seconds<=0) { errno=EINVAL; return -1; }
 struct addrinfo hints, *list, *p; memset(&hints,0,sizeof hints);
 hints.ai_socktype=SOCK_STREAM; hints.ai_family=AF_UNSPEC;
 if(getaddrinfo(host,port,&hints,&list)!=0) { errno=EHOSTUNREACH; return -1; }
 int fd=-1, saved=ECONNREFUSED;
 for(p=list;p;p=p->ai_next) {
  fd=socket(p->ai_family,p->ai_socktype,p->ai_protocol); if(fd<0) { saved=errno; continue; }
  struct timeval tv={seconds,0};
  if(setsockopt(fd,SOL_SOCKET,SO_SNDTIMEO,&tv,sizeof tv)==0 && setsockopt(fd,SOL_SOCKET,SO_RCVTIMEO,&tv,sizeof tv)==0 && connect(fd,p->ai_addr,p->ai_addrlen)==0) break;
  saved=errno; close(fd); fd=-1;
 }
 freeaddrinfo(list); if(fd<0) errno=saved; return fd;
}
int aether_send(int fd,const void *payload,size_t length) {
 if(length>UINT32_MAX) { errno=EMSGSIZE; return -1; }
 if(length && !payload) { errno=EINVAL; return -1; }
 unsigned char h[11]={0xae,0x74,0x45,0x52,0,1,1,0,0,0,0};
 uint32_t n=(uint32_t)length; h[7]=(unsigned char)(n>>24); h[8]=(unsigned char)(n>>16); h[9]=(unsigned char)(n>>8); h[10]=(unsigned char)n;
 if(send_all(fd,h,sizeof h)<0) return -1;
 return send_all(fd,payload,length);
}
int aether_close(int fd) { return close(fd); }
