package main

import (
 "encoding/json"
 "flag"
 "fmt"
 "os"

 "github.com/XenoTechWorks/Astelyn/internal/compataudit"
 "github.com/XenoTechWorks/Astelyn/internal/darling"
 "github.com/XenoTechWorks/Astelyn/internal/inspection"
)

func runCompatAudit(args []string) int {
 flags:=flag.NewFlagSet("compat-audit",flag.ContinueOnError)
 flags.SetOutput(os.Stderr)
 root:=flags.String("root",darling.SystemRoot,"Darling runtime root to examine")
 comparison:=flags.String("compare-root","","optional second runtime root (does not execute it)")
 jsonOut:=flags.Bool("json",false,"output JSON report")
 if err:=flags.Parse(args);err!=nil{return 2}
 if flags.NArg()!=1 {
  fmt.Fprintln(os.Stderr,"Usage: astelyn compat-audit [--json] [--root /path] [--compare-root /path] <RobloxPlayer.zip>")
  return 2
 }
 inspected,err:=inspection.InspectZip(flags.Arg(0))
 if err!=nil {fmt.Fprintln(os.Stderr,"Astelyn:",err);return 1}
 deps:=inspection.AnalyzeDependencies(inspected)
 result,err:=compataudit.Analyze(deps,*root,*comparison)
 if err!=nil {fmt.Fprintln(os.Stderr,"Astelyn:",err);return 1}
 if *jsonOut {if err:=json.NewEncoder(os.Stdout).Encode(result);err!=nil{fmt.Fprintln(os.Stderr,err);return 1};return 0}
 fmt.Printf("Direct framework imports: %d\nPrimary: %s (exact=%d alternate=%d missing=%d)\n",len(result.Rows),result.PrimaryRoot,result.Primary.Exact,result.Primary.Alternate,result.Primary.Missing)
 if result.Comparison!=nil {fmt.Printf("Comparison: %s (exact=%d alternate=%d missing=%d)\n",result.ComparisonRoot,result.Comparison.Exact,result.Comparison.Alternate,result.Comparison.Missing)}
 for _,row:=range result.Rows {
  fmt.Printf("  %-10s %s",row.PrimaryState,row.Import)
  if result.Comparison!=nil{fmt.Printf("  [comparison: %s]",row.ComparisonState)}
  fmt.Println()
 }
 for _,n:=range result.Notes{fmt.Println("Note:",n)}
 return 0
}
