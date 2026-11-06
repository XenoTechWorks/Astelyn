/* An SDK-free source example for a macOS (Mach-O) cross-compiler.
 * This C program CALLS Objective-C runtime APIs, without relying on Foundation.
 * It is NOT directly buildable by an ordinary Linux ELF compiler.
 * Build with a Darwin-targeting toolchain and appropriate system libraries.
 */
#include <stdio.h>
#include <stddef.h>

extern void *objc_getClass(const char *name);
extern void *sel_registerName(const char *name);

int main(void) {
    void *selector = sel_registerName("description");
    void *base_class = objc_getClass("NSObject");
    if (selector == NULL) {
        fputs("objc runtime selector registration failed\n", stderr);
        return 1;
    }
    if (base_class == NULL)
        puts("NSObject unavailable (Foundation not loaded or implemented)");
    else
        puts("NSObject class found");
    puts("ASTELYN_OBJC_OK");
    return 0;
}
