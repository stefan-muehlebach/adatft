//
// Dies ist ein Versuch, mittels C-Programm direkt auf das Adafruit
// TFT-Display zu schreiben. Die Grundlage fuer dieses Programm war tftcp.
//

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <errno.h>
#include <sys/mman.h>
#include <sys/ioctl.h>
#include <linux/spi/spidev.h>
#include <bcm_host.h>
#include "ili9341.h"

#define ILI9341_NOP        0x00
#define ILI9341_SWRESET    0x01
#define ILI9341_RDDID      0x04
#define ILI9341_RDDST      0x09

#define ILI9341_SLPIN      0x10
#define ILI9341_SLPOUT     0x11
#define ILI9341_PTLON      0x12
#define ILI9341_NORON      0x13

#define ILI9341_RDMODE     0x0A
#define ILI9341_RDMADCTL   0x0B
#define ILI9341_RDPIXFMT   0x0C
#define ILI9341_RDIMGFMT   0x0D
#define ILI9341_RDSELFDIAG 0x0F

#define ILI9341_INVOFF     0x20
#define ILI9341_INVON      0x21
#define ILI9341_GAMMASET   0x26
#define ILI9341_DISPOFF    0x28
#define ILI9341_DISPON     0x29

#define ILI9341_CASET      0x2A
#define ILI9341_PASET      0x2B
#define ILI9341_RAMWR      0x2C
#define ILI9341_RAMRD      0x2E

#define ILI9341_PTLAR      0x30
#define ILI9341_MADCTL     0x36
#define ILI9341_VSCRSADD   0x37
#define ILI9341_PIXFMT     0x3A

#define ILI9341_CABC       0x55

#define ILI9341_FRMCTR1    0xB1
#define ILI9341_FRMCTR2    0xB2
#define ILI9341_FRMCTR3    0xB3
#define ILI9341_INVCTR     0xB4
#define ILI9341_DFUNCTR    0xB6

#define ILI9341_PWCTR1     0xC0
#define ILI9341_PWCTR2     0xC1
#define ILI9341_PWCTR3     0xC2
#define ILI9341_PWCTR4     0xC3
#define ILI9341_PWCTR5     0xC4
#define ILI9341_VMCTR1     0xC5
#define ILI9341_VMCTR2     0xC7

#define ILI9341_PWCTRLA    0xCB
#define ILI9341_PWCTRLB    0xCF

#define ILI9341_RDID1      0xDA
#define ILI9341_RDID2      0xDB
#define ILI9341_RDID3      0xDC
#define ILI9341_RDID4      0xDD

#define ILI9341_GMCTRP1    0xE0
#define ILI9341_GMCTRN1    0xE1
#define ILI9341_DRVTICTRLA 0xE8
#define ILI9341_DRVTICTRLB 0xEA
#define ILI9341_PWOSEQCTR  0xED

#define ILI9341_3G         0xF2
#define ILI9341_PMPRTCTR   0xF7

#define MADCTL_MY          0x80
#define MADCTL_MX          0x40
#define MADCTL_MV          0x20
#define MADCTL_ML          0x10
#define MADCTL_RGB         0x00
#define MADCTL_BGR         0x08
#define MADCTL_MH          0x04

#define DC_PIN             25
#define BITRATE            80000000
#define BUFFER_SIZE        4096

#define SPI_MODE  SPI_MODE_0
#define FPS       30

#define PI1_BCM2708_PERI_BASE 0x20000000
#define PI1_GPIO_BASE         (PI1_BCM2708_PERI_BASE + 0x200000)
#define PI2_BCM2708_PERI_BASE 0x3F000000
#define PI2_GPIO_BASE         (PI2_BCM2708_PERI_BASE + 0x200000)
#define BLOCK_SIZE            (4*1024)
#define INP_GPIO(g)          *(gpio+((g)/10)) &= ~(7<<(((g)%10)*3))
#define OUT_GPIO(g)          *(gpio+((g)/10)) |=  (1<<(((g)%10)*3))

#define ERR_BUF_LEN 256

static volatile unsigned
    *gpio = NULL,
    *gpioSet,
    *gpioClr;

uint32_t resetMask, dcMask;

char errStr[ERR_BUF_LEN];
char *ILI_errStr = errStr;

static struct spi_ioc_transfer
  cmd = {
    .rx_buf = 0,
    .delay_usecs = 0,
    .bits_per_word = 8,
    .pad = 0,
    .speed_hz = BITRATE,
    .tx_nbits = 0,
    .rx_nbits = 0,
    .cs_change = 0 },
  dat = {
    .rx_buf = 0,
    .delay_usecs = 0,
    .bits_per_word = 8,
    .len = BUFFER_SIZE,
    .pad = 0,
    .speed_hz = BITRATE,
    .tx_nbits = 0,
    .rx_nbits = 0,
    .cs_change = 0 };

int ILI_open() {
    int fd;
    if ((fd = open("/dev/spidev0.0", O_WRONLY | O_NONBLOCK)) < 0) {
        sprintf(ILI_errStr, "Can't open /dev/spidev0.0: %s", strerror(errno));
        return -1;
    }
    uint8_t  mode  = SPI_MODE;
    uint32_t speed = BITRATE;
    if (ioctl(fd, SPI_IOC_WR_MODE, &mode) < 0) {
        sprintf(ILI_errStr, "Can't set SPI mode: %s", strerror(errno));
        return -1;
    }
    if (ioctl(fd, SPI_IOC_WR_MAX_SPEED_HZ, &speed) < 0) {
        sprintf(ILI_errStr, "Can't set SPI speed: %s", strerror(errno));
        return -1;
    }
    return fd;
}

int ILI_init(int fd) {
    int gfd;
    if ((gfd = open("/dev/mem", O_RDWR | O_SYNC)) < 0) {
        sprintf(ILI_errStr, "Can't open /dev/mem: %s", strerror(errno));
        return -1;
    }
    gpio = (volatile unsigned *) mmap(
        NULL,
        BLOCK_SIZE,
        PROT_READ | PROT_WRITE,
        MAP_SHARED,
        gfd,
        PI2_GPIO_BASE);
    close(gfd);
    if (gpio == MAP_FAILED) {
        sprintf(ILI_errStr, "Can't map(): %s", strerror(errno));
        return -1;
    }
    gpioSet = &gpio[7];
    gpioClr = &gpio[10];
    dcMask = 1 << DC_PIN;

    INP_GPIO(DC_PIN);
    OUT_GPIO(DC_PIN);

    ILI_cmd(fd, ILI9341_SWRESET);	// Reset before close

    ILI_cmd(fd, ILI9341_MADCTL);	// Memory Access Control
    ILI_data8(fd, 0x28);

    ILI_cmd(fd, ILI9341_PIXFMT);	// Pixel Format Set
#ifdef PIXFMT_565
    ILI_data8(fd, 0x05);
#else
    ILI_data8(fd, 0x06);
#endif

    ILI_cmd(fd, ILI9341_3G);		// Disable 3G (Gamma)
    ILI_data8(fd, 0x02);

    ILI_cmd(fd, ILI9341_GAMMASET);	// Gamma Set
    ILI_data8(fd, 0x01);

    uint8_t posGamma[] = {
        0x0f, 0x31, 0x28, 0x0c, 0x0e, 0x08, 0x4e,
        0xf1,
        0x37, 0x07, 0x10, 0x03, 0x0e, 0x09, 0x00
    };
    ILI_cmd(fd, ILI9341_GMCTRP1);	// Positive Gamma Correction
    ILI_data(fd, posGamma, 15);

    uint8_t negGamma[] = {
        0x00, 0x0e, 0x14, 0x03, 0x11, 0x07, 0x31,
        0xc1,
        0x48, 0x08, 0x0f, 0x0c, 0x31, 0x36, 0x0f
    };
    ILI_cmd(fd, ILI9341_GMCTRN1);	// Negative Gamma Correction
    ILI_data(fd, negGamma, 15);

    ILI_cmd(fd, ILI9341_SLPOUT);	// Exit Sleep
    usleep(120000);

    ILI_cmd(fd, ILI9341_DISPON);	// Display On
    usleep(120000);

    return 0;
}

void ILI_close (int fd) {
    ILI_cmd(fd, ILI9341_SWRESET);	// Reset before close
    close(fd);
}

//-----------------------------------------------------------------------------

void ILI_cmd(int fd, uint8_t c) {
    *gpioClr = dcMask;
    cmd.tx_buf = (unsigned long)&c;
    cmd.len = 1;
    (void)ioctl(fd, SPI_IOC_MESSAGE(1), &cmd);
}

void ILI_data8(int fd, uint8_t value) {
    *gpioSet = dcMask;
    cmd.tx_buf = (unsigned long)&value;
    cmd.len = 1;
    (void)ioctl(fd, SPI_IOC_MESSAGE(1), &cmd);
}

void ILI_data16(int fd, uint16_t value) {
    uint8_t foo[2];
    *gpioSet = dcMask;
    foo[0] = value >>  8;
    foo[1] = value;
    cmd.tx_buf = (unsigned long)&foo;
    cmd.len = 2;
    (void)ioctl(fd, SPI_IOC_MESSAGE(1), &cmd);
}

void ILI_data32(int fd, uint32_t value) {
    uint8_t foo[4];
    *gpioSet = dcMask;
    foo[0] = value >> 24;
    foo[1] = value >> 16;
    foo[2] = value >>  8;
    foo[3] = value;
    cmd.tx_buf = (unsigned long)&foo;
    cmd.len = 4;
    (void)ioctl(fd, SPI_IOC_MESSAGE(1), &cmd);
}

void ILI_data(int fd, uint8_t *buf, unsigned len) {
    uint32_t bytesRemaining;
    *gpioSet = dcMask;
    bytesRemaining = len;
    dat.tx_buf = (uint32_t)buf;
    while (bytesRemaining > 0) {
        dat.len = (bytesRemaining > BUFFER_SIZE) ? BUFFER_SIZE : bytesRemaining;
        (void)ioctl(fd, SPI_IOC_MESSAGE(1), &dat);
        bytesRemaining -= dat.len;
        dat.tx_buf += dat.len;
    }
}

//-----------------------------------------------------------------------------
//
//
struct TFT {
    int fd;
    double dCol, dRow, xMin, yMax;
    RgbType foreColor, backColor;
    RgbType *pixelBuffer;
};

#ifdef PIXFMT_565
    RgbType TFT_BLACK   = { .r=0x00, .gh=0x0, .gl=0x0, .b=0x00 };
    RgbType TFT_WHITE   = { .r=0x1f, .gh=0x7, .gl=0x7, .b=0x1f };
    RgbType TFT_RED     = { .r=0x1f, .gh=0x0, .gl=0x0, .b=0x00 };
    RgbType TFT_GREEN   = { .r=0x00, .gh=0x7, .gl=0x7, .b=0x00 };
    RgbType TFT_BLUE    = { .r=0x00, .gh=0x0, .gl=0x0, .b=0x1f };
    RgbType TFT_YELLOW  = { .r=0x1f, .gh=0x7, .gl=0x7, .b=0x00 };
    RgbType TFT_CYAN    = { .r=0x00, .gh=0x7, .gl=0x7, .b=0x1f };
    RgbType TFT_MAGENTA = { .r=0x1f, .gh=0x0, .gl=0x0, .b=0x1f };
    RgbType TFT_GRAY    = { .r=0x0f, .gh=0x3, .gl=0x7, .b=0x0f };
#else
    RgbType TFT_BLACK   = { 0x00, 0x00, 0x00 };
    RgbType TFT_WHITE   = { 0x3f, 0x3f, 0x3f };
    RgbType TFT_RED     = { 0x3f, 0x00, 0x00 };
    RgbType TFT_GREEN   = { 0x00, 0x3f, 0x00 };
    RgbType TFT_BLUE    = { 0x00, 0x00, 0x3f };
    RgbType TFT_YELLOW  = { 0x3f, 0x3f, 0x00 };
    RgbType TFT_CYAN    = { 0x00, 0x3f, 0x3f };
    RgbType TFT_MAGENTA = { 0x3f, 0x00, 0x3f };
    RgbType TFT_GRAY    = { 0x1f, 0x1f, 0x1f };
#endif

TFT TFT_new () {
    TFT tft;

    tft = malloc(sizeof(*tft));
    tft->dCol = +1.0;
    tft->dRow = -1.0;
    tft->xMin =  0.0;
    tft->yMax =  0.0;
    tft->foreColor = TFT_WHITE;
    tft->backColor = TFT_BLACK;
    tft->pixelBuffer = calloc(TFT_HEIGHT*TFT_WIDTH, sizeof(RgbType));

    if ((tft->fd = ILI_open()) < 0) {
        return NULL;
    }
    if (ILI_init(tft->fd) < 0) {
        return NULL;
    }

    return tft;
}

void TFT_free (TFT tft) {
    close(tft->fd);
    free(tft->pixelBuffer);
    free(tft);
}

void TFT_clear(TFT tft) {
    for (int i=0; i<TFT_WIDTH * TFT_HEIGHT; i++) {
        tft->pixelBuffer[i] = tft->backColor;
    }
}

void TFT_show(TFT tft) {
    ILI_cmd(tft->fd, ILI9341_CASET);
    ILI_data32(tft->fd, TFT_WIDTH-1);
    ILI_cmd(tft->fd, ILI9341_PASET);
    ILI_data32(tft->fd, TFT_HEIGHT-1);
    ILI_cmd(tft->fd, ILI9341_RAMWR);
    ILI_data(tft->fd, (uint8_t *) tft->pixelBuffer,
            TFT_WIDTH*TFT_HEIGHT*sizeof(RgbType));
}

void TFT_setDisplay(TFT tft, int onoff) {
    if (onoff) {
        ILI_cmd(tft->fd, ILI9341_DISPON);
    } else {
        ILI_cmd(tft->fd, ILI9341_DISPOFF);
    }
}

void TFT_setInversion(TFT tft, int onoff) {
    if (onoff) {
        ILI_cmd(tft->fd, ILI9341_INVON);
    } else {
        ILI_cmd(tft->fd, ILI9341_INVOFF);
    }
}

void TFT_setBackColor (TFT tft, RgbType color) {
    tft->backColor = color;
}

void TFT_setForeColor (TFT tft, RgbType color) {
    tft->foreColor = color;
}

//-----------------------------------------------------------------------------
//
// Drawing functions
//

inline void setColorPixel(TFT tft, int x, int y, RgbType color) {
    if (x < 0 || x >= TFT_WIDTH || y < 0 || y >= TFT_HEIGHT) {
        return;
    }
    tft->pixelBuffer[TFT_WIDTH*y + x] = color;
}

inline void setPixel(TFT tft, int x, int y) {
    setColorPixel(tft, x, y, tft->foreColor);
}

/*
#define setPixel(tft, x, y) \
    if (x < 0 || x >= TFT_WIDTH || y < 0 || y >= TFT_HEIGHT) { \
        return; \
    } \
    tft->pixelBuffer[TFT_WIDTH*y + x] = tft->foreColor; \
*/

void TFT_drawPoint(TFT tft, int x, int y) {
    setPixel(tft, x, y);
}

void TFT_drawPointC (TFT tft, int x, int y, RgbType color) {
    setColorPixel(tft, x, y, color);
}

void TFT_drawLine(TFT tft, int x1, int y1, int x2, int y2) {
    int dx =  abs(x2 - x1), sx = (x1 < x2) ? 1 : -1;
    int dy = -abs(y2 - y1), sy = (y1 < y2) ? 1 : -1;
    int err = dx + dy, e2;

//    printf("dx: %d, dy: %d\n", dx, dy);
//    printf("sx: %d, sy: %d\n", sx, sy);
    while (1) {
//        printf("x: %2d, y: %2d, err: %d\n", x1, y1, err);
        setPixel(tft, x1, y1);
        if (x1 == x2 && y1 == y2) break;
        e2 = 2 * err;
        if (e2 > dy) { err += dy; x1 += sx; }
        if (e2 < dx) { err += dx; y1 += sy; }
    }
}

void TFT_drawRect(TFT tft, int x1, int y1, int x2, int y2) {
    for (int x=x1; x<=x2; x++) {
        setPixel(tft, x, y1);
        setPixel(tft, x, y2);
    }
    for (int y=y1; y<=y2; y++) {
        setPixel(tft, x1, y);
        setPixel(tft, x2, y);
    }
}

void TFT_drawEllipse(TFT tft, int xm, int ym, int a, int b) {
    int dx = 0, dy = b;
    long a2 = a*a, b2 = b*b;
    long err = b2-(2*b-1)*a2, e2;

    do {
        setPixel(tft, xm + dx, ym + dy);
        setPixel(tft, xm - dx, ym + dy);
        setPixel(tft, xm - dx, ym - dy);
        setPixel(tft, xm + dx, ym - dy);
        e2 = 2*err;
        if (e2 <  (2*dx+1) * b2) { ++dx; err += (2*dx+1) * b2; }
        if (e2 > -(2*dy-1) * a2) { --dy; err -= (2*dy-1) * a2; }
    } while (dy >= 0);
    while (dx++ < a) {
        setPixel(tft, xm+dx, ym);
        setPixel(tft, xm-dx, ym);
    }
}

void TFT_setDisplayRange(TFT tft, double xMin, double xMax,
        double yMin, double yMax) {
    tft->xMin = xMin;
    tft->yMax = yMax;
    tft->dCol = TFT_WIDTH / (xMax - xMin);
    tft->dRow = TFT_HEIGHT / (yMax - yMin);
}

inline int toCol(TFT tft, double x) {
    return (int)((x - tft->xMin) * tft->dCol);
}

inline int toRow(TFT tft, double y) {
    return (int)((tft->yMax - y) * tft->dRow);
}

void TFT_drawPointM(TFT tft, double x, double y) {
    setPixel(tft, toCol(tft, x), toRow(tft, y));
}

void TFT_drawLineM(TFT tft, double x1, double y1, double x2, double y2) {
    TFT_drawLine(tft, toCol(tft, x1), toRow(tft, y1),
            toCol(tft, x2), toRow(tft, y2));
}

void TFT_drawRectM(TFT tft, double x1, double y1, double x2, double y2) {
    TFT_drawRect(tft, toCol(tft, x1), toRow(tft, y1),
            toCol(tft, x2), toRow(tft, y2));
}

void TFT_drawEllipseM(TFT tft, double xm, double ym, double a, double b) {
    TFT_drawEllipse(tft, toCol(tft, xm), toRow(tft, ym),
            a*tft->dCol, b*tft->dRow);
}


