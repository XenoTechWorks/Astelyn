package guicheck
import("os";"path/filepath";"testing")
func TestMissing(t *testing.T){r:=t.TempDir();report,err:=Analyze(r);if err!=nil{t.Fatal(err)};if len(report.Components)!=7||report.CocoaWindowReady{t.Fatalf("unexpected report: %+v",report)}}
func TestRoot(t *testing.T){if _,err:=Analyze("relative");err==nil{t.Fatal("relative root accepted")}}
func TestEscape(t *testing.T){root:=t.TempDir();outside:=t.TempDir();folder:=filepath.Join(root,"usr","lib");if err:=os.MkdirAll(folder,0755);err!=nil{t.Fatal(err)};if err:=os.Symlink(outside,filepath.Join(folder,"libobjc.A.dylib"));err!=nil{t.Fatal(err)};_,err:=within(root,"/usr/lib/libobjc.A.dylib");if err==nil{t.Fatal("accepted escaping symlink")}}
