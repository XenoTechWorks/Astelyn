/* Astelyn v0.2.7: native Linux display test; NOT a Cocoa implementation. */
#include <SDL.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(int argc, char **argv) {
    unsigned seconds = 12;
    if (argc == 3 && strcmp(argv[1], "--seconds") == 0) {
        char *end = NULL;
        unsigned long n = strtoul(argv[2], &end, 10);
        if (!end || *end != '\0' || n < 1 || n > 300) {
            fprintf(stderr, "seconds must be an integer from 1 to 300\n");
            return 2;
        }
        seconds = (unsigned)n;
    } else if (argc != 1) {
        fprintf(stderr, "usage: astelyn-host-window [--seconds 1..300]\n");
        return 2;
    }
    if (SDL_Init(SDL_INIT_VIDEO) != 0) {
        fprintf(stderr, "SDL_Init: %s\n", SDL_GetError());
        return 1;
    }
    SDL_Window *window = SDL_CreateWindow("Astelyn - Linux Window Backend Test",
        SDL_WINDOWPOS_CENTERED, SDL_WINDOWPOS_CENTERED, 720, 420,
        SDL_WINDOW_SHOWN | SDL_WINDOW_RESIZABLE);
    if (!window) {
        fprintf(stderr, "SDL_CreateWindow: %s\n", SDL_GetError());
        SDL_Quit();
        return 1;
    }
    const char *driver = SDL_GetCurrentVideoDriver();
    printf("ASTELYN_HOST_WINDOW_CREATED backend=%s\n", driver ? driver : "unknown");
    fflush(stdout);
    SDL_Renderer *renderer = SDL_CreateRenderer(window, -1, SDL_RENDERER_SOFTWARE);
    if (!renderer) {
        fprintf(stderr, "SDL_CreateRenderer: %s\n", SDL_GetError());
        SDL_DestroyWindow(window);
        SDL_Quit();
        return 1;
    }
    Uint32 start = SDL_GetTicks();
    int running = 1;
    while (running && (Uint32)(SDL_GetTicks() - start) < seconds * 1000U) {
        SDL_Event event;
        while (SDL_PollEvent(&event)) {
            if (event.type == SDL_QUIT ||
                (event.type == SDL_KEYDOWN && event.key.keysym.sym == SDLK_ESCAPE)) {
                running = 0;
            }
        }
        SDL_SetRenderDrawColor(renderer, 23, 30, 53, 255);
        SDL_RenderClear(renderer);
        int w, h;
        SDL_GetRendererOutputSize(renderer, &w, &h);
        SDL_Rect bar = {w / 6, h / 2 - 14, w * 2 / 3, 28};
        SDL_SetRenderDrawColor(renderer, 89, 185, 219, 255);
        SDL_RenderFillRect(renderer, &bar);
        SDL_RenderPresent(renderer);
        SDL_Delay(16);
    }
    SDL_DestroyRenderer(renderer);
    SDL_DestroyWindow(window);
    SDL_Quit();
    puts("ASTELYN_HOST_WINDOW_CLOSED");
    return 0;
}
