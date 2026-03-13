#include "webview.h"

#include <stdlib.h>
#include <stdint.h>

void _webviewDispatchGoCallback(uintptr_t);
void _webviewBindingGoCallback(const char *, const char *, uintptr_t);

static void _webview_dispatch_cb(webview_t w, void *arg) {
    _webviewDispatchGoCallback((uintptr_t)arg);
}

static void _webview_binding_cb(const char *id, const char *req, void *arg) {
    _webviewBindingGoCallback(id, req, (uintptr_t)arg);
}

void CgoWebViewDispatch(webview_t w, uintptr_t arg) {
    webview_dispatch(w, _webview_dispatch_cb, (void *)arg);
}

void CgoWebViewBind(webview_t w, const char *name, uintptr_t arg) {
    webview_bind(w, name, _webview_binding_cb, (void *)arg);
}

void CgoWebViewUnbind(webview_t w, const char *name) {
    webview_unbind(w, name);
}
