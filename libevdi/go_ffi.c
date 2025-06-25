#include <stdio.h>
#include <stdarg.h>

#include "evdi_lib.h"

// Events

extern void goDPMSHandler(int, void*);
extern void goModeChangedHandler(struct evdi_mode, void*);
extern void goUpdateReadyHandler(int, void*);
extern void goCRTCStateHandler(int, void*);
extern void goCursorSetHandler(struct evdi_cursor_set, void*);
extern void goCursorMoveHandler(struct evdi_cursor_move, void*);
extern void goDDCCIDataHandler(struct evdi_ddcci_data, void*);

void dpmsHandler(int dpms_mode, void *user_data) {
	goDPMSHandler(dpms_mode, user_data);
}

void modeChangedHandler(struct evdi_mode mode, void *user_data) {
	goModeChangedHandler(mode, user_data);
}

void updateReadyHandler(int crtc_id, void *user_data) {
	goUpdateReadyHandler(crtc_id, user_data);
}

void crtcStateHandler(int crtc_state, void *user_data) {
	goCRTCStateHandler(crtc_state, user_data);
}

void cursorSetHandler(struct evdi_cursor_set cursor_set, void *user_data) {
	goCursorSetHandler(cursor_set, user_data);
}

void cursorMoveHandler(struct evdi_cursor_move cursor_move, void *user_data) {
	goCursorMoveHandler(cursor_move, user_data);
}

void ddcciDataHandler(struct evdi_ddcci_data ddcci_data, void *user_data) {
	goDDCCIDataHandler(ddcci_data, user_data);
}

// Logging

extern void goLoggerHandler(char*);

void loggerHandler(void* user_data, const char *fmt, ...) {
    char* dynBuf;

    va_list args;
    va_start(args, fmt);
    vasprintf(&dynBuf, fmt, args);
    va_end(args);

    goLoggerHandler(dynBuf);
}

void loggerInit() {
    struct evdi_logging evdiLogger = { loggerHandler };

    evdi_set_logging(evdiLogger);
}
