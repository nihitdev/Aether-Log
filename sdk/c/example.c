#include "aether.h"
#include <stdio.h>
#include <string.h>
int main(int argc,char **argv) {
 const char *host=argc>1?argv[1]:"127.0.0.1";
 const char *port=argc>2?argv[2]:"8080";
 const char *record=argc>3?argv[3]:"hello from C";
 int fd=aether_connect(host,port,5); if(fd<0) { perror("connect"); return 1; }
 int result=aether_send(fd,record,strlen(record)); if(result<0) perror("send");
 aether_close(fd); return result<0;
}
