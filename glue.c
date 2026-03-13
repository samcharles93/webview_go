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

webview_error_t CgoWebViewDispatch(webview_t w, uintptr_t arg) {
    return webview_dispatch(w, _webview_dispatch_cb, (void *)arg);
}

webview_error_t CgoWebViewBind(webview_t w, const char *name, uintptr_t arg) {
    return webview_bind(w, name, _webview_binding_cb, (void *)arg);
}

webview_error_t CgoWebViewUnbind(webview_t w, const char *name) {
    return webview_unbind(w, name);
}
