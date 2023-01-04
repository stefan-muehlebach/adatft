all: libili9341.so

CFLAGS=-Ofast -fomit-frame-pointer -DPIXFMT_565

libili9341.so: ili9341.c ili9341.h
	gcc -c ${CFLAGS} -Wall -Werror -fpic -o ili9341.o ili9341.c
	gcc -shared -o libili9341.so ili9341.o

