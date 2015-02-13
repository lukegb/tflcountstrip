#include "application.h"
//#include "spark_disable_wlan.h" (for faster local debugging only)
#include "neopixel/neopixel.h"

#define PIXEL_COUNT 150
#define PIXEL_PIN D2
#define PIXEL_TYPE WS2812B

#define BUS_STOP (PIXEL_COUNT / 2)
#define UPDATE_INTERVAL 100 // minimum amount of time between updates

#define MAX_BUSSES 100

Adafruit_NeoPixel strip = Adafruit_NeoPixel(PIXEL_COUNT, PIXEL_PIN, PIXEL_TYPE);

#define NO_SYNC -1
#define SYNC_TIMEOUT 30
int lastSync;

const uint32_t COLOURS[] = {
    strip.Color(0, 0, 0), // No bus
    strip.Color(255, 0, 0), // Too late
    strip.Color(255, 165, 0), // RUN!
    strip.Color(0, 255, 0), // Just right
    strip.Color(75, 0, 130), // Wait a bit?
    strip.Color(255, 255, 255), // It's a bus stop
    
    strip.Color(0, 0, 255), // error
    strip.Color(0, 0, 255), // error
    strip.Color(0, 0, 255), // error
    strip.Color(0, 0, 255), // error
    strip.Color(255, 0, 255), // sync failed
};

char buf[192];

int charToInt(char n) {
    switch (n) {
        case '0': return 0;
        case '1': return 1;
        case '2': return 2;
        case '3': return 3;
        case '4': return 4;
        case '5': return 5;
        case '6': return 6;
        case '7': return 7;
        case '8': return 8;
        case '9': return 9;
    }
    return 10;
}

#define REDRAW1 0x1
#define REDRAW2 0x2
#define REDRAW3 0x4
#define REDRAWALL (REDRAW1 | REDRAW2 | REDRAW3)
int redrawData = 0;

#define SPLITPOINT 63

int updateData(int startat, int redrawer, String args) {
    char tmpBuf[SPLITPOINT];
    args.toCharArray(tmpBuf, SPLITPOINT);
    memcpy(&buf[startat], tmpBuf, SPLITPOINT-1);
    
    redrawData |= redrawer;
    if ((redrawData & REDRAWALL) == REDRAWALL) {
        lastSync = Time.now();
        redrawData = 0;
        for (int i = 0; i < PIXEL_COUNT; i++) {
            int col = charToInt(buf[i]);
            strip.setPixelColor(i, COLOURS[col]);
        }
        strip.show();
    }
    
    return 0;
}

int update1(String args) { return updateData(0, REDRAW1, args); }
int update2(String args) { return updateData(SPLITPOINT - 1, REDRAW2, args); }
int update3(String args) { return updateData(SPLITPOINT*2 - 2, REDRAW3, args); }

void setup() {
    lastSync = NO_SYNC;
    
    strip.begin();
    
    for (int i = 0; i < PIXEL_COUNT; i++) {
        buf[i] = '9';
    }
    
    strip.show(); // Initialize all pixels
    
    Spark.function("push1", update1);
    Spark.function("push2", update2);
    Spark.function("push3", update3);
}

void loop() {
    int now = Time.now();
    if (lastSync < now - SYNC_TIMEOUT) {
        for (int i = 0; i < PIXEL_COUNT; i++) {
            strip.setPixelColor(i, COLOURS[11]);
        }
        strip.show();
    }
}

