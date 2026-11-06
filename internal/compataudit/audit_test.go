package compataudit
import (
 "os"
 "path/filepath"
 "testing"
 "github.com/XenoTechWorks/Astelyn/internal/inspection"
)
func TestCompareRoots(t *testing.T){
 a:=t.TempDir();b:=t.TempDir()
 rel:="System/Library/Frameworks/CoreML.framework/Versions/A/CoreML"
 full:=filepath.Join(b,rel)
 if err:=os.MkdirAll(filepath.Dir(full),0755);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(full,[]byte("sample-not-a-framework"),0644);err!=nil{t.Fatal(err)}
 dep:=inspection.DependencyReport{Archive:"test.zip",Dependencies:[]inspection.Dependency{{Name:"/"+rel,Category:"apple-framework",UsedBy:[]string{"RobloxPlayer"}}}}
 r,err:=Analyze(dep,a,b);if err!=nil{t.Fatal(err)}
 if r.Primary.Missing!=1||r.Comparison.Exact!=1||len(r.Rows)!=1 {t.Fatalf("bad counts: %+v",r)}
 if r.Rows[0].PrimaryState!="missing"||r.Rows[0].ComparisonState!="exact"{t.Fatalf("bad states: %+v",r.Rows)}
 if len(r.Notes)==0||len(r.Sources)!=3{t.Fatal("documentation missing")}
}
func TestInvalidRoot(t *testing.T){
 _,err:=Analyze(inspection.DependencyReport{},"relative","")
 if err==nil{t.Fatal("expected invalid root error")}
}
func TestNoComparison(t *testing.T){
 root:=t.TempDir();r,err:=Analyze(inspection.DependencyReport{},root,"")
 if err!=nil||r.Comparison!=nil||len(r.Rows)!=0{t.Fatalf("unexpected: %+v %v",r,err)}
}
