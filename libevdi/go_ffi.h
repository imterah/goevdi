#include "evdi_lib.h"

void dpmsHandler(int dpms_mode, void *user_data);
void modeChangedHandler(struct evdi_mode mode, void *user_data);
void updateReadyHandler(int crtc_id, void *user_data);
void crtcStateHandler(int crtc_state, void *user_data);
void cursorSetHandler(struct evdi_cursor_set cursor_set, void *user_data);
void cursorMoveHandler(struct evdi_cursor_move cursor_move, void *user_data);
void ddcciDataHandler(struct evdi_ddcci_data ddcci_data, void *user_data);
void loggerHandler(void* user_data, const char *fmt, ...);
void loggerInit();

extern void goDPMSHandler(int, void*);
extern void goModeChangedHandler(struct evdi_mode, void*);
extern void goUpdateReadyHandler(int, void*);
extern void goCRTCStateHandler(int, void*);
extern void goCursorSetHandler(struct evdi_cursor_set, void*);
extern void goCursorMoveHandler(struct evdi_cursor_move, void*);
extern void goDDCCIDataHandler(struct evdi_ddcci_data, void*);
extern void goLoggerHandler(char*);
