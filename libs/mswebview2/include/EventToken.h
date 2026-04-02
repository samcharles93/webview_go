#ifndef WEBVIEW_COMPAT_EVENTTOKEN_H
#define WEBVIEW_COMPAT_EVENTTOKEN_H
#ifdef _WIN32

// Compatibility header for Windows GNU-family toolchains that do not ship
// EventToken.h with the expected casing or do not ship it at all.
//
// Source model: https://github.com/webview/webview compatibility/mingw/include/EventToken.h

#ifndef __eventtoken_h__

#ifdef __cplusplus
#include <cstdint>
#else
#include <stdint.h>
#endif

typedef struct EventRegistrationToken {
  int64_t value;
} EventRegistrationToken;

#endif // __eventtoken_h__

#endif // _WIN32
#endif // WEBVIEW_COMPAT_EVENTTOKEN_H
