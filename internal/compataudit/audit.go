// Package compataudit compares direct macOS framework imports against installed
// Darling system roots. It does not modify the runtime or assess API support.
package compataudit

import (
 "errors"
 "fmt"
 "os"
 "path/filepath"
 "sort"
 "strings"

 "github.com/XenoTechWorks/Astelyn/internal/frameworkresolve"
 "github.com/XenoTechWorks/Astelyn/internal/inspection"
)

type RootCount struct {
 Exact int `json:"exact"`
 Alternate int `json:"alternate"`
 Missing int `json:"missing"`
}
type Row struct {
 Import string `json:"import"`
 Framework string `json:"framework"`
 UsedBy []string `json:"used_by"`
 PrimaryState string `json:"primary_state"`
 PrimaryHostPath string `json:"primary_host_path,omitempty"`
 ComparisonState string `json:"comparison_state,omitempty"`
 ComparisonHostPath string `json:"comparison_host_path,omitempty"`
}
type Source struct {
 Project string `json:"project"`
 URL string `json:"url"`
 DocumentedApproach string `json:"documented_approach"`
 Verification string `json:"verification"`
}
type Report struct {
 Archive string `json:"archive"`
 PrimaryRoot string `json:"primary_root"`
 ComparisonRoot string `json:"comparison_root,omitempty"`
 Primary RootCount `json:"primary_counts"`
 Comparison *RootCount `json:"comparison_counts,omitempty"`
 Rows []Row `json:"frameworks"`
 Sources []Source `json:"related_projects"`
 Notes []string `json:"notes"`
}

func validateRoot(root string) error {
 if !filepath.IsAbs(root) || strings.ContainsAny(root, "\x00\r\n") { return errors.New("runtime root must be an absolute path without control characters") }
 info, err := os.Stat(root)
 if err != nil { return fmt.Errorf("runtime root %q: %w", root, err) }
 if !info.IsDir() { return fmt.Errorf("runtime root %q is not a directory", root) }
 return nil
}
func count(c *RootCount, state string) {
 switch state {case "exact": c.Exact++; case "alternate": c.Alternate++; default: c.Missing++}
}

// Analyze uses only direct Mach-O imports from an existing ZIP inspection.
// A file present in a comparison root is NOT proof of usable Cocoa/CoreML APIs.
func Analyze(deps inspection.DependencyReport, primary, comparison string) (Report,error) {
 if err:=validateRoot(primary);err!=nil{return Report{},err}
 if comparison!="" {if err:=validateRoot(comparison);err!=nil{return Report{},err}}
 p:=frameworkresolve.Analyze(deps,primary)
 r:=Report{Archive:deps.Archive, PrimaryRoot:primary, ComparisonRoot:comparison, Rows:[]Row{}, Sources:[]Source{
  {Project:"Quadra",URL:"https://github.com/QuadraHQ/Quadra",DocumentedApproach:"Uses additional Darwin/Darling runtime fixes and libQuadraShims.dylib; offers regular Darling or CiderDarling.",Verification:"Documentation describes runtime modifications; it does not establish an implementation of any particular missing framework."},
  {Project:"CiderDarling",URL:"https://github.com/QuadraHQ/CiderDarling",DocumentedApproach:"Modified Darling runtime with additional compatibility changes, including OpenGL-oriented work.",Verification:"Source location for investigation; no framework availability inferred."},
  {Project:"Mac O' Blox",URL:"https://github.com/narezy/MacOBlox",DocumentedApproach:"Original project that informed Quadra.",Verification:"Historical reference only; check current licensing and functionality independently."},
 }, Notes:[]string{
  "Only direct framework imports in Mach-O binaries recognized by Astelyn's ZIP inspector are included; dynamic dlopen and indirect imports are not covered.",
  "Framework file presence is not evidence of API, ABI, rendering, or Roblox compatibility.",
  "This audit reads files and does not install, inject, replace, or fabricate libraries.",
  "CoreML-specific handling in other projects is not verified by this report.",
 }}
 for _,e:=range p.Entries {
  row:=Row{Import:e.Import,Framework:e.Framework,UsedBy:append([]string{},e.UsedBy...),PrimaryState:e.State,PrimaryHostPath:e.HostPath}
  count(&r.Primary,e.State)
  r.Rows=append(r.Rows,row)
 }
 if comparison!="" {
  c:=frameworkresolve.Analyze(deps,comparison)
  counts:=&RootCount{}
  others:=map[string]frameworkresolve.Entry{}
  for _,e:=range c.Entries {others[e.Import]=e}
  for i:=range r.Rows {
   if e,ok:=others[r.Rows[i].Import];ok {
    r.Rows[i].ComparisonState=e.State
    r.Rows[i].ComparisonHostPath=e.HostPath
    count(counts,e.State)
   } else {r.Rows[i].ComparisonState="missing";counts.Missing++}
  }
  r.Comparison=counts
 }
 sort.Slice(r.Rows,func(i,j int)bool{return r.Rows[i].Import<r.Rows[j].Import})
 return r,nil
}
