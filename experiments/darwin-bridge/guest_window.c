/* Astelyn v0.2.9: Darwin Mach-O guest window REQUEST, not an NSWindow.
 * Calls libSystem's puts(). The Go host validates these lines before opening
 * its native SDL2 window. No Cocoa, networking or macOS SDK is required.
 */
extern int puts(const char *);
int main(void) {
    puts("ASTELYN_GUEST_WINDOW_V1");
    puts("TITLE Astelyn v0.2.9 - Darwin guest request");
    puts("SIZE 840 520");
    return 0;
}
