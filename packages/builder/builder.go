// builder
package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"net/http"
	"runtime"
	"io"
	"path/filepath"
	"os/exec"
	"strings"
	"reflect"
	"time"
)

const (
//	GITPATH = `github.com/democratic-coin/dcoin-go`
	GITPATH = `github.com/DayLightProject/go-daylight`
)

var (
    options Settings
)

type Settings struct {
	Branch    string   // Branch name 
	GitRoot   string   // https://github.com/democratic-coin/dcoin-go
	TempPath  string   // Temporary path
	OutFile   string   // Output dcoin executable file
	GoPath    string   // GOPATH
	BinData   string   // Full path to go-bindata 
	BinDebug  string   // Specify "true" for debug option
	RunAfter  [][]string   
	Arch      string   // Custom GOARCH
	Replace   string   // Replace string
	Skip      string   // d - download, z - unzip, s - static.go, b - build, a - make package
}

func exit( err error ) {  
	if err != nil {
		fmt.Println( err )
	}
	fmt.Println( `Press Enter to exit...`)
    fmt.Scanln( )
	if err != nil {
		os.Exit(1)
	}
}

func download( zfile string ) ( destfile string ) {
	srcfile := fmt.Sprintf("%s/archive/%s", options.GitRoot, zfile )
	destfile = filepath.Join( options.TempPath, zfile )
	if strings.IndexRune( options.Skip, 'd' ) >= 0 {
		return
	}
	fmt.Println(`Downloading `, srcfile )
	out, err := os.Create( destfile )
  	if err != nil  {
  		exit( err )
    }
  	defer out.Close()
	check := http.Client{
                 CheckRedirect: func(r *http.Request, via []*http.Request) error {
                         r.URL.Opaque = r.URL.Path
                         return nil },
     }
    resp, err := check.Get( srcfile )
  	if err != nil {
  		exit(err)
  	}
	defer resp.Body.Close()
	  _, err = io.Copy(out, resp.Body)
  	if err != nil  {
  		exit(err)
  	}
	fmt.Println(`Downloaded successfully` )
	return
}

func  extract(f *zip.File) error {
    rc, err := f.Open()
    if err != nil {
        return err
    }
    defer rc.Close()
	fname := f.Name[strings.IndexRune( f.Name, '/' ) + 1:]
    path := filepath.Join( options.GoPath, `src`, GITPATH, fname )
	fmt.Println(`Decompressing`, fname )
    if f.FileInfo().IsDir() {
        return os.MkdirAll(path, f.Mode())
    } else {
        f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            return err
        }
        defer f.Close()
		
        _, err = io.Copy(f, rc)
        if err != nil {
            return err
        }
    }
    return nil
}

func BytesInfoHeader(size int, filename string) (*zip.FileHeader, error) {
	fh := &zip.FileHeader{
		Name:               filename,
		UncompressedSize64: uint64(size),
		UncompressedSize:   uint32(size),
		Method:             zip.Deflate,
	}
	fh.SetModTime(time.Now())
	//   fh.SetMode(fi.Mode())
	return fh, nil
}

func main() {
	var ( settings map[string]Settings
	)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		exit(err)
	}
	params, err := ioutil.ReadFile(filepath.Join(dir, `builder_day.json`))
	if err != nil {
		exit(err)
	}
	if err = json.Unmarshal(params, &settings); err != nil {
		exit(err)
	}
	options = settings[`default`]
	if len(os.Args ) > 1 {
		if cmdopt, found := settings[ os.Args[1]]; found {
			r := reflect.ValueOf(cmdopt)
			for i:=0; i < r.NumField(); i++ {
				if reflect.ValueOf(&cmdopt).Elem().Type().Field(i).Name == `RunAfter` {
					continue
				}
				val := r.Field(i).String()
				if len( val ) > 0 {
					ro := reflect.ValueOf(&options)
					ro.Elem().Field(i).SetString( val )
				}
			}
			if len( cmdopt.RunAfter ) > 0 {
				options.RunAfter = cmdopt.RunAfter
			}
		} else {
			exit( fmt.Errorf( `Cannot find %s settings`, os.Args[1]))
		}
	}
	fmt.Println( options )
	srcPath := filepath.Join( options.GoPath, `src`, GITPATH )
	if err = os.MkdirAll( srcPath, 0755); err != nil {
		exit(err)
	}
	fmt.Println( `Destination Path`, srcPath )
	zfile := options.Branch + `.zip`
	srcfile := download( zfile )
	
	if strings.IndexRune( options.Skip, 'z' ) < 0 {
		z, err := zip.OpenReader(srcfile)
		if err != nil {
			exit( err )
		}
		defer z.Close()
		if _, err := os.Stat( filepath.Join( srcPath, "dcoinwindows.go")); err == nil {
			fmt.Println(`Removing `, srcPath )		
			if err = os.RemoveAll( srcPath); err!=nil {
				exit(err)
			}
		}
		
		for _, f := range z.File {
			if err = extract( f ); err != nil {
				exit( err )
			}
		}
	}

	if err = os.Chdir( srcPath ); err != nil {
		exit( err )
	}
	if strings.IndexRune( options.Skip, 's' ) < 0 {
		fmt.Println(`Creating static.go`)
		args := []string{ `-o=packages/static/static.go`, `-pkg=static` }
		if options.BinDebug == `true` {
			args = append( args, `-debug=true`)
		}
		cmd := exec.Command( options.BinData, append( args, `static/...` )...)
		if err := cmd.Run(); err != nil {
			exit( err )
		}
	}
	
	fmt.Println( options )
	if len( options.Replace) > 0 {
		fmt.Println(`Replace`, options.Replace )
		static := filepath.Join( srcPath, "packages", "static","static.go")
		in,err := ioutil.ReadFile( static )
		if err != nil {
			fmt.Println("Error in", err.Error())
		}
		out := strings.Replace(string(in), `"`+options.Replace+`static`, `"static`, -1 )
		err = ioutil.WriteFile( static, []byte(out), 0644 )
		if err != nil {
			fmt.Println("Error out", err.Error())
		}
	}
	
	if strings.IndexRune( options.Skip, 'b' ) < 0 {
		fmt.Println(`Compiling dcoin.go`)
		if err = os.MkdirAll( filepath.Dir(options.OutFile), 0755); err != nil {
			exit(err)
		}
		os.Setenv(`GOPATH`, options.GoPath )
		if len( options.Arch ) > 0 {
			os.Setenv(`GOARCH`, options.Arch )
		}
		args := []string{ `build`, `-o`, options.OutFile, `-ldflags` }
		if runtime.GOOS == `windows` {
			args = append( args, `-H windowsgui`)
		}
		cmd := exec.Command( `go`, append( args, GITPATH )... )
		if err = cmd.Run(); err != nil {
			exit( err )
		}
	}
	zipfile := `daylight.zip`
	switch runtime.GOOS {
		case `windows`:
			if runtime.GOARCH == `386` {
				zipfile = `daylight_win32.zip` 
			} else {
				zipfile = `daylight_win64.zip` 
			}
	}
	zipname := filepath.Join(filepath.Dir( filepath.Dir( options.OutFile )), zipfile )
	fmt.Println(`Compressing`, zipname)
	
	zipf, err := os.Create(zipname)
	if err != nil {
		exit( err )
	}
	z := zip.NewWriter(zipf)
	var out []byte
	if out, err = ioutil.ReadFile( options.OutFile ); err != nil {
		exit(err)
	}
	header, _ := BytesInfoHeader(len(out), filepath.Base( options.OutFile ))
	f,_ := z.CreateHeader(header)
	f.Write(out)
	z.Close()
	zipf.Close()
	
	if len(options.RunAfter) > 0 && strings.IndexRune( options.Skip, 'a' ) < 0 {
		fmt.Println(`Run at the end`)
		for _, icmd := range options.RunAfter {
			if len(icmd) > 0 && len( icmd[0]) > 0 {
				fmt.Println(`Executing`, icmd[0], `...` )
				out, err := exec.Command( icmd[0], icmd[1:]... ).Output()
				if err != nil {
					exit( err )
				}
				fmt.Println( string(out) )
			}
		}
	}
	exit(nil)
}
