package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	//viewPathTemplates glob template map
	viewPathTemplates = make(map[string]*template.Template)

	//locker
	rwLock sync.RWMutex

	//defaultViewPath
	defaultViewPath = "views"

	//templateFuncMap
	templateFuncMap = make(template.FuncMap)

	//utf8 text/html ContentType
	htmlContentType = []string{"text/html; charset=utf-8"}
)

//Delims Delims
type Delims struct {
	Left  string
	Right string
}

//HTMLRender a simple html render
type HTMLRender struct {
	Template *template.Template
	Name     string
	Data     interface{}
}

// init templateFuncMap
// internal template func
// {{.title | str2html}}
func init() {
	templateFuncMap["str2html"] = Str2html
	templateFuncMap["html2str"] = HTML2str
	templateFuncMap["date_time_format"] = DateTimeFormat
	templateFuncMap["date"] = DateFormat
	templateFuncMap["int_date_time_format"] = IntDateTimeFormat
	templateFuncMap["int_date_time"] = IntDateTime
	templateFuncMap["int_date"] = IntDate
	templateFuncMap["substr"] = Substr
	templateFuncMap["assets_js"] = AssetsJs
	templateFuncMap["assets_css"] = AssetsCSS
}

func (r HTMLRender) Instance(name string, data interface{}) Render {
	var buf bytes.Buffer
	execViewPathTemplate(&buf, name, data)
	return HTMLRender{
		Template: viewPathTemplates[name],
		Data:     data,
		Name:     name,
	}
}

// Render (HTML) executes template and writes its result with custom ContentType for response.
func (r HTMLRender) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	if r.Name == "" {
		return r.Template.Execute(w, r.Data)
	}
	return r.Template.ExecuteTemplate(w, r.Name, r.Data)
}

// WriteContentType (HTML) writes HTML ContentType.
func (r HTMLRender) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, htmlContentType)
}

//execViewPathTemplate
func execViewPathTemplate(wr io.Writer, name string, data interface{}) error {
	rwLock.RLock()
	defer rwLock.RUnlock()
	if t, ok := viewPathTemplates[name]; ok {
		var err error
		if t.Lookup(name) != nil {
			err = t.ExecuteTemplate(wr, name, data)
		} else {
			err = t.Execute(wr, data)
		}
		return err
	}

	panic("cat not find template file in path: " + defaultViewPath + "/" + name)
}

type templateFile struct {
	root  string
	files map[string][]string
}

func (tf *templateFile) visit(paths string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
		return nil
	}

	replace := strings.NewReplacer("\\", "/")
	file := strings.TrimLeft(replace.Replace(paths[len(tf.root):]), "/")
	subDir := filepath.Dir(file)

	tf.files[subDir] = append(tf.files[subDir], file)
	return nil
}

//BuildTemplate init template
func BuildTemplate(dir string, funcMap template.FuncMap, delims Delims, files ...string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.New("dir opens failed")
	}
	for key, item := range funcMap {
		templateFuncMap[key] = item
	}
	tff := &templateFile{
		root:  dir,
		files: make(map[string][]string),
	}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		return tff.visit(path, f, err)
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
		return err
	}

	buildAllFiles := len(files) == 0
	for _, v := range tff.files {
		for _, file := range v {
			if buildAllFiles || InSlice(file, files) {
				rwLock.Lock()
				var t *template.Template
				t, err = getTemplate(tff.root, file, delims, v...)
				if err != nil {
					log.Printf("parse template err: %v", err)
					rwLock.Unlock()
					return err
				}
				viewPathTemplates[file] = t
				rwLock.Unlock()
			}
		}
	}

	return nil
}

func getTemplate(root, file string, delims Delims, others ...string) (t *template.Template, err error) {
	t = template.New(file).Funcs(templateFuncMap).Delims(delims.Left, delims.Right)
	var subMods [][]string
	t, subMods, err = getTplDeep(root, file, "", t)
	if err != nil {
		return nil, err
	}
	t, err = _getTemplate(t, root, subMods, others...)

	if err != nil {
		return nil, err
	}
	return
}

func _getTemplate(t0 *template.Template, root string, subMods [][]string, others ...string) (t *template.Template, err error) {
	t = t0
	for _, m := range subMods {
		if len(m) == 2 {
			tpl := t.Lookup(m[1])
			if tpl != nil {
				continue
			}
			//first check filename
			for _, otherFile := range others {
				if otherFile == m[1] {
					var subMods1 [][]string
					t, subMods1, err = getTplDeep(root, otherFile, "", t)
					if err != nil {
						log.Printf("template parse file err: %v", err)
					} else if len(subMods1) > 0 {
						t, err = _getTemplate(t, root, subMods1, others...)
					}
					break
				}
			}
			//second check define
			for _, otherFile := range others {
				var data []byte
				fileAbsPath := filepath.Join(root, otherFile)
				data, err = ioutil.ReadFile(fileAbsPath)
				if err != nil {
					continue
				}
				reg := regexp.MustCompile("{{[ ]*define[ ]+\"([^\"]+)\"")
				allSub := reg.FindAllStringSubmatch(string(data), -1)
				for _, sub := range allSub {
					if len(sub) == 2 && sub[1] == m[1] {
						var subMods1 [][]string
						t, subMods1, err = getTplDeep(root, otherFile, "", t)
						if err != nil {
							fmt.Printf("template parse file err: %v", err)
						} else if len(subMods1) > 0 {
							t, err = _getTemplate(t, root, subMods1, others...)
						}
						break
					}
				}
			}
		}

	}
	return
}

func getTplDeep(root, file, parent string, t *template.Template) (*template.Template, [][]string, error) {
	var fileAbsPath string
	var rParent string
	if filepath.HasPrefix(file, "../") {
		rParent = filepath.Join(filepath.Dir(parent), file)
		fileAbsPath = filepath.Join(root, filepath.Dir(parent), file)
	} else {
		rParent = file
		fileAbsPath = filepath.Join(root, file)
	}
	if e := FileExists(fileAbsPath); !e {
		panic("can't find template file:" + file)
	}
	data, err := ioutil.ReadFile(fileAbsPath)
	if err != nil {
		return nil, [][]string{}, err
	}
	t, err = t.New(file).Parse(string(data))
	if err != nil {
		return nil, [][]string{}, err
	}
	reg := regexp.MustCompile("{{[ ]*template[ ]+\"([^\"]+)\"")
	allSub := reg.FindAllStringSubmatch(string(data), -1)
	for _, m := range allSub {
		if len(m) == 2 {
			tl := t.Lookup(m[1])
			if tl != nil {
				continue
			}
			_, _, err = getTplDeep(root, m[1], rParent, t)
			if err != nil {
				return nil, [][]string{}, err
			}
		}
	}
	return t, allSub, nil
}

// InSlice checks given string in string slice or not.
func InSlice(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

// FileExists reports whether the named file or directory exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
