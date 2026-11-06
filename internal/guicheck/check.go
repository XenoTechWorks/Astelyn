// Package guicheck reports whether Darling has the on-disk prerequisites
// for experimenting with a Cocoa window. It does not claim GUI support.
package guicheck

import (
 "debug/macho"
 "errors"
 "fmt"
 "os"
 "path/filepath"
 "strings"
)

type Component struct {
 Name string `json:"name"`
 GuestPath string `json:"guest_path"`
 Present bool `json:"present"`
 MachO bool `json:"macho"`
 Error string `json:"error,omitempty"`
}
type Report struct {
 DarlingRoot string `json:"darling_root"`
 Components []Component `json:"components"`
 CocoaWindowReady bool `json:"cocoa_window_ready"`
 Note string `json:"note"`
}
var targets = []struct{name, path string}{
 {"Objective-C runtime", "/usr/lib/libobjc.A.dylib"},
 {"Darwin C runtime", "/usr/lib/libSystem.B.dylib"},
 {"AppKit", "/System/Library/Frameworks/AppKit.framework/Versions/C/AppKit"},
 {"Foundation", "/System/Library/Frameworks/Foundation.framework/Versions/C/Foundation"},
 {"CoreGraphics", "/System/Library/Frameworks/CoreGraphics.framework/Versions/A/CoreGraphics"},
 {"QuartzCore", "/System/Library/Frameworks/QuartzCore.framework/Versions/A/QuartzCore"},
 {"Metal", "/System/Library/Frameworks/Metal.framework/Versions/A/Metal"},
}
func within(root, guest string) (string, error) {
 if !filepath.IsAbs(root) || !filepath.IsAbs(guest) || strings.ContainsAny(guest, "\x00\n\r") {return "",errors.New("absolute root and guest paths required")}
 resolvedRoot,err:=filepath.EvalSymlinks(root);if err!=nil{return "",err}
 candidate:=filepath.Join(root,strings.TrimPrefix(guest,"/"))
 resolved,err:=filepath.EvalSymlinks(candidate);if err!=nil{return "",err}
 rel,err:=filepath.Rel(resolvedRoot,resolved);if err!=nil{return "",err}
 if rel==".." || strings.HasPrefix(rel,".."+string(filepath.Separator)) {return "",errors.New("symlink points outside Darling root")}
 return resolved,nil
}
func Analyze(root string) (Report,error) {
 st,err:=os.Stat(root);if err!=nil{return Report{},err};if !st.IsDir() || !filepath.IsAbs(root){return Report{},fmt.Errorf("Darling root must be an absolute directory")}
 report:=Report{DarlingRoot:root,Note:"Static checks only. Objective-C classes must be checked at runtime. Framework presence does not prove Cocoa window support."}
 for _,target:=range targets {
  item:=Component{Name:target.name,GuestPath:target.path}
  path,e:=within(root,target.path)
  if e!=nil {if !errors.Is(e,os.ErrNotExist){item.Error=e.Error()}} else {
   f,e:=macho.Open(path)
   if e!=nil {item.Error=e.Error()} else {item.Present=true;item.MachO=true;f.Close()}
  }
  report.Components=append(report.Components,item)
 }
 report.CocoaWindowReady=report.Components[2].MachO && report.Components[3].MachO && report.Components[4].MachO
 return report,nil
}
