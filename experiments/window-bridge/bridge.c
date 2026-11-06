/* Astelyn v0.2.8: Linux SDL2 host-window IPC experiment, NOT Cocoa emulation.
 * Line protocol on stdin: PING / SIZE w h / TITLE text / QUIT.
 * Line events on stdout. Do not treat this as a network-facing interface.
 */
#include <SDL.h>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/select.h>
#include <unistd.h>

static void line(const char *s) { puts(s); fflush(stdout); }
static void process(const char *s, SDL_Window *win, int *running) {
 if (strcmp(s,"PING")==0) {line("PONG");return;}
 if (strcmp(s,"QUIT")==0) {*running=0;line("QUIT");return;}
 int w,h;char tail;
 if (sscanf(s,"SIZE %d %d %c",&w,&h,&tail)==2 && w>=200 && w<=1920 && h>=150 && h<=1080) {
  SDL_SetWindowSize(win,w,h);return;
 }
 if (strncmp(s,"TITLE ",6)==0) {
  const char *t=s+6;size_t n=strlen(t);
  int valid=(n>=1 && n<=100);
  for(size_t i=0;i<n;i++){
   unsigned char ch=(unsigned char)t[i];
   if(ch<32 || ch>126 || ch=='|' || ch=='\\') {valid=0;break;}
  }
  if(valid) {
   SDL_SetWindowTitle(win,t);return;
  }
 }
 line("ERROR unsupported-command");
}
int main(void) {
 if (SDL_Init(SDL_INIT_VIDEO)!=0){fprintf(stderr,"SDL_Init: %s\n",SDL_GetError());return 1;}
 SDL_Window *win=SDL_CreateWindow("Astelyn window bridge",SDL_WINDOWPOS_CENTERED,SDL_WINDOWPOS_CENTERED,720,420,SDL_WINDOW_SHOWN|SDL_WINDOW_RESIZABLE);
 if (!win){fprintf(stderr,"SDL_CreateWindow: %s\n",SDL_GetError());SDL_Quit();return 1;}
 SDL_Renderer *ren=SDL_CreateRenderer(win,-1,SDL_RENDERER_SOFTWARE);
 if (!ren){fprintf(stderr,"SDL_CreateRenderer: %s\n",SDL_GetError());SDL_DestroyWindow(win);SDL_Quit();return 1;}
 const char *driver=SDL_GetCurrentVideoDriver();printf("DRIVER %s\n",driver?driver:"unknown");line("READY");
 int running=1;char input[256];size_t len=0;int stdin_open=1;
 Uint32 began=SDL_GetTicks();
 while(running && SDL_GetTicks()-began<30000U){
  SDL_Event e;while(SDL_PollEvent(&e)){
   if(e.type==SDL_QUIT){running=0;break;}
   if(e.type==SDL_KEYDOWN&&e.key.keysym.sym==SDLK_ESCAPE){line("KEY escape");running=0;break;}
   if(e.type==SDL_WINDOWEVENT&&e.window.event==SDL_WINDOWEVENT_SIZE_CHANGED){printf("RESIZED %d %d\n",e.window.data1,e.window.data2);fflush(stdout);}
  }
  if(stdin_open){
   fd_set rfds;FD_ZERO(&rfds);FD_SET(STDIN_FILENO,&rfds);struct timeval timeout={0,0};
   int sel=select(STDIN_FILENO+1,&rfds,NULL,NULL,&timeout);
   if(sel>0&&FD_ISSET(STDIN_FILENO,&rfds)){
    char ch;ssize_t n=read(STDIN_FILENO,&ch,1);
    if(n==0){stdin_open=0;running=0;}
    else if(n>0){
     if(ch=='\n'){input[len]='\0';process(input,win,&running);len=0;}
     else if(ch=='\r'){}else if(len<sizeof(input)-1){input[len++]=ch;}
     else {line("ERROR line-too-long");len=0;}
    }
   }
  }
  SDL_SetRenderDrawColor(ren,23,30,53,255);SDL_RenderClear(ren);
  int w,h;SDL_GetRendererOutputSize(ren,&w,&h);
  SDL_Rect bar={w/8,h/2-12,w*3/4,24};SDL_SetRenderDrawColor(ren,89,185,219,255);SDL_RenderFillRect(ren,&bar);
  SDL_RenderPresent(ren);SDL_Delay(16);
 }
 line("CLOSED");SDL_DestroyRenderer(ren);SDL_DestroyWindow(win);SDL_Quit();return 0;
}
