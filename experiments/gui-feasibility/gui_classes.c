/* Astelyn v0.2.6: SDK-free Darling Objective-C class availability probe.
 * This program does NOT open a window. It checks classes and selectors only.
 */
extern void *objc_getClass(const char *name);
extern void *sel_registerName(const char *name);
extern int puts(const char *s);

static void check(const char *name, const char *present, const char *missing) {
    puts(objc_getClass(name) ? present : missing);
}
int main(void) {
    if (!sel_registerName("alloc") || !sel_registerName("init")) {
        puts("ASTELYN_GUI_SELECTOR_FAILURE"); return 2;
    }
    check("NSObject", "CLASS NSObject present", "CLASS NSObject missing");
    check("NSApplication", "CLASS NSApplication present", "CLASS NSApplication missing");
    check("NSWindow", "CLASS NSWindow present", "CLASS NSWindow missing");
    check("NSView", "CLASS NSView present", "CLASS NSView missing");
    check("CALayer", "CLASS CALayer present", "CLASS CALayer missing");
    check("CAMetalLayer", "CLASS CAMetalLayer present", "CLASS CAMetalLayer missing");
    puts("ASTELYN_GUI_CLASS_PROBE_OK");
    return 0;
}
