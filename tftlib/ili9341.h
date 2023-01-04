/*-----------------------------------------------------------------------------
 *
 * ili9341.h --
 *
 *     Header file der Library, welche den Zugang zum 320x240 TFT Display
 *     von Adafruit vereinfacht.
 *
 * History
 *     20221215  Umbau zu ILI9341, Auftrennung der HW-Fuktionen und der
 *               Zeichenbefehle.
 *     20221101  Initiale Version
 *
 *-----------------------------------------------------------------------------
 */

#include <stdint.h>

#ifndef ILI9341_H
#define ILI9341_H

// Low level commands for the ILI9341
//
// Open, initialize and close the connection to the TFT
//
extern int  ILI_open   ();
extern int  ILI_init   (int fd);
extern void ILI_close  (int fd);

// Functions for communication with the ILI controller via SPI
//
extern void ILI_cmd    (int fd, uint8_t cmd);
extern void ILI_data8  (int fd, uint8_t value);
extern void ILI_data16 (int fd, uint16_t value);
extern void ILI_data32 (int fd, uint32_t value);
extern void ILI_data   (int fd, uint8_t *buf, unsigned len);

extern char *ILI_errStr;

//-----------------------------------------------------------------------------
//
// High level commands and structures for the ILI9341
//
#define TFT_WIDTH  320
#define TFT_HEIGHT 240

typedef struct TFT *TFT;

// Type of a pixel/color and some color constants.
//
#ifdef PIXFMT_565
    typedef struct RgbType {
        uint8_t gh:3, r:5;
        uint8_t b:5, gl:3;
    } RgbType;

    #define SET_RED(C,V)       C.r=V;
    #define SET_GREEN(C,V) \
        C.gh = (V >> 3); \
        C.gl = (V & 0x07);
    #define SET_BLUE(C,V)      C.b=V;
    #define SET_COLOR(C,R,G,B) C.r=R; C.gh=(G>>3); C.gl=(G&0x07); C.b=B;


    #define MAX_RED_VALUE   31
    #define MAX_GREEN_VALUE 63
    #define MAX_BLUE_VALUE  31
#else
    typedef struct RgbType {
        uint8_t :2, r:6;
        uint8_t :2, g:6;
        uint8_t :2, b:6;
    } RgbType;

    #define SET_RED(C,V)       C.r=V;
    #define SET_GREEN(C,V)     C.g=V;
    #define SET_BLUE(C,V)      C.b=V;
    #define SET_COLOR(C,R,G,B) C.r=R; C.g=G; C.b=B;

    #define MAX_RED_VALUE   63
    #define MAX_GREEN_VALUE 63
    #define MAX_BLUE_VALUE  63
#endif

extern RgbType TFT_BLACK;
extern RgbType TFT_WHITE;
extern RgbType TFT_RED;
extern RgbType TFT_GREEN;
extern RgbType TFT_BLUE;
extern RgbType TFT_YELLOW;
extern RgbType TFT_CYAN;
extern RgbType TFT_MAGENTA;
extern RgbType TFT_GRAY;

// Initialize and discard TFT structures.
//
extern TFT  TFT_new  ();
extern void TFT_free (TFT tft);

// Clear and display the content of the back buffer.
//
extern void TFT_clear (TFT tft);
extern void TFT_show  (TFT tft);

extern void TFT_setDisplay (TFT tft, int onoff);
extern void TFT_setInversion (TFT tft, int onoff);

// Color setting functions.
//
extern void TFT_setBackColor (TFT tft, RgbType color);
extern void TFT_setForeColor (TFT tft, RgbType color);

// Drawing functions for...
//
// Pixels
//
extern void TFT_drawPoint (TFT tft, int x, int y);
extern void TFT_drawPointC (TFT tft, int x, int y, RgbType color);

// Lines
//
extern void TFT_drawLine  (TFT tft, int x1, int y1, int x2, int y2);

// Rectangles
//
extern void TFT_drawRect (TFT tft, int x1, int y1, int x2, int y2);

// Circles/Ellipses
//
extern void TFT_drawEllipse (TFT tft, int xm, int ym, int a, int b);

// Drawing with mathematical coordinates.
//
extern void TFT_setDisplayRange (TFT tft, double xMin, double xMax,
        double yMin, double yMax);

extern void TFT_drawPointM (TFT tft, double x, double y);
extern void TFT_drawLineM  (TFT tft, double x1, double y1,
        double x2, double y2);
extern void TFT_drawRectM (TFT tft, double x1, double y1,
        double x2, double y2);
extern void TFT_drawEllipseM (TFT tft, double xm, double ym,
        double a, double b);

#endif

