/*
 * Astelyn v0.2.5 — Objective-C runtime smoke test.
 * No macOS SDK headers required.
 */

extern void *objc_getClass(const char *name);
extern void *sel_registerName(const char *name);
extern int puts(const char *text);

int main(void) {
    void *selector = sel_registerName("description");
    void *base_class = objc_getClass("NSObject");

    if (!selector) {
        puts("Objective-C selector registration failed");
        return 1;
    }

    if (base_class)
        puts("NSObject class found");
    else
        puts("NSObject unavailable");

    puts("ASTELYN_OBJC_OK");
    return 0;
}
