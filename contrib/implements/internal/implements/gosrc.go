package implements

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
)

type goFunction struct {
	shortName       string
	fullName        string
	comment         string
	implementsCFunc string
	callsCFunc      string
	isDeprecated    bool
	isPreview       bool

	endPos token.Pos
}

type visitor struct {
	currentFunc *goFunction

	callMap    map[string]*goFunction
	docMap     map[string]*goFunction
	deprecated []*goFunction
	preview    []*goFunction
}

func newVisitor() *visitor {
	return &visitor{
		callMap:    map[string]*goFunction{},
		docMap:     map[string]*goFunction{},
		deprecated: []*goFunction{},
		preview:    []*goFunction{},
	}
}

func funcDeclFullName(fdec *ast.FuncDecl) string {
	if fdec.Recv == nil {
		return fdec.Name.Name
	}
	if len(fdec.Recv.List) != 1 {
		return fdec.Name.Name
	}
	typeName := "UNKNOWN!"
	switch t := fdec.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		typeName = t.X.(*ast.Ident).Name
	case *ast.Ident:
		typeName = t.Name
	}
	return fmt.Sprintf("%s.%s", typeName, fdec.Name.Name)
}

func readDocComment(fdec *ast.FuncDecl, gfunc *goFunction) {
	gfunc.comment = fdec.Doc.Text()
	lines := strings.Split(gfunc.comment, "\n")
	for i := range lines {
		if strings.Contains(lines[i], "DEPRECATED") {
			gfunc.isDeprecated = true
			logger.Printf("marked deprecated: %s\n", fdec.Name.Name)
		}
		if strings.Contains(lines[i], "PREVIEW") {
			gfunc.isPreview = true
			logger.Printf("marked preview: %s\n", fdec.Name.Name)
		}

		if lines[i] == "Implements:" {
			cfunc := cfuncFromComment(lines[i+1])
			if cfunc == "" {
				return
			}
			gfunc.implementsCFunc = cfunc
			logger.Printf("implements c function %s: %s\n", fdec.Name.Name, cfunc)
		}
	}
}

func (v *visitor) checkCalled(s *ast.SelectorExpr) {
	ident, ok := s.X.(*ast.Ident)
	if !ok {
		return
	}
	if "C" == ident.String() {
		v.callMap[s.Sel.String()] = v.currentFunc
		logger.Printf("updated %s in call map\n", s.Sel.String())
	}
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch {
	case node == nil:
		return nil
	case v.currentFunc == nil:
	case node.Pos() > v.currentFunc.endPos:
		logger.Printf("left function %v\n", v.currentFunc.shortName)
		v.currentFunc = nil
	}

	switch n := node.(type) {
	case *ast.File:
		v.currentFunc = nil
		return v
	case *ast.FuncDecl:
		logger.Printf("checking function: %v\n", n.Name.Name)
		gfunc := &goFunction{
			shortName: n.Name.Name,
			fullName:  funcDeclFullName(n),
			endPos:    n.End(),
		}
		readDocComment(n, gfunc)
		v.currentFunc = gfunc
		if gfunc.isDeprecated {
			v.deprecated = append(v.deprecated, gfunc)
			logger.Printf("rem1 %v\n", v.deprecated)
		}
		if gfunc.isPreview {
			v.preview = append(v.preview, gfunc)
			logger.Printf("rem2 %v\n", v.preview)
		}
		return v
	case *ast.CallExpr:
		if v.currentFunc == nil {
			return nil
		}
		if s, ok := n.Fun.(*ast.SelectorExpr); ok {
			v.checkCalled(s)
		}
	}
	if v.currentFunc != nil {
		return v
	}
	return nil
}

func cfuncFromComment(ctext string) string {
	m := regexp.MustCompile(` ([a-zA-Z0-9_]+)\(`).FindAllSubmatch([]byte(ctext), 1)
	if len(m) < 1 {
		return ""
	}
	return string(m[0][1])
}

// CephGoFunctions will look for C functions called by the code code and
// update the found functions for the package within the inspector.
func CephGoFunctions(source, packageName string, ii *Inspector) error {
	p, err := build.Import("./"+packageName, source, 0)
	if err != nil {
		return err
	}

	toCheck := []string{}
	toCheck = append(toCheck, p.GoFiles...)
	toCheck = append(toCheck, p.CgoFiles...)
	for _, fname := range toCheck {
		logger.Printf("Reading go file: %v\n", fname)
		src, err := ioutil.ReadFile(path.Join(p.Dir, fname))
		if err != nil {
			return err
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(
			fset,
			fname,
			src,
			parser.ParseComments|parser.AllErrors)
		if err != nil {
			return err
		}
		ast.Walk(ii.visitor, f)
	}
	return nil
}
