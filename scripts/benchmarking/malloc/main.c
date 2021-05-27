#include <stdlib.h>
#include <unistd.h>
#include <stdio.h>
#include <sys/mman.h>

int main(int argc, char *argv[])
{
	setlinebuf(stdout);
	const int bytes = 100 * 1024 * 1024;
	void* mem;
	while(1){
		sleep(1);
		char* mem = malloc (bytes);
		if (mem == NULL){
			printf("Failed to allocate memory\n");
			continue;
		}
		mlock (mem, bytes);
		size_t page_size=getpagesize();
		for (size_t i = 0; i < bytes; i++) {
			mem[i] = 0;
		}
		printf("Wrote %d bytes\n", bytes);
	}
	return 0;
}
