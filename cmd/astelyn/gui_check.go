package main
import("encoding/json";"flag";"fmt";"os";"github.com/XenoTechWorks/Astelyn/internal/guicheck";"github.com/XenoTechWorks/Astelyn/internal/darling")
func runGUICheck(args []string) int {
 flags:=flag.NewFlagSet("gui-check",flag.ContinueOnError)
 flags.SetOutput(os.Stderr)
 root:=flags.String("root",darling.SystemRoot,"Darling filesystem root")
 jsonOut:=flags.Bool("json",false,"output JSON")
 if err:=flags.Parse(args);err!=nil{return 2}
 if flags.NArg()!=0 {fmt.Fprintln(os.Stderr,"Usage: astelyn gui-check [--json] [--root path]");return 2}
 report,err:=guicheck.Analyze(*root);if err!=nil{fmt.Fprintln(os.Stderr,"Astelyn:",err);return 1}
 if *jsonOut {if err:=json.NewEncoder(os.Stdout).Encode(report);err!=nil{return 1};return 0}
 for _,entry:=range report.Components {status:="missing";if entry.MachO{status="Mach-O present"};if entry.Error!=""{status="error: "+entry.Error};fmt.Printf("%-22s %s\n",entry.Name,status)}
 fmt.Println("Cocoa window prerequisites present:",report.CocoaWindowReady)
 fmt.Println(report.Note)
 return 0
}
