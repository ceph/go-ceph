package implements

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
)

type visitor struct {
	inFunction *ast.FuncDecl

	callMap    map[string]string
	docMap     map[string]string
	deprecated map[string]string
	preview    map[string]string
}

func newVisitor() *visitor {
	return &visitor{
		callMap:    map[string]string{},
		docMap:     map[string]string{},
		deprecated: map[string]string{},
		preview:    map[string]string{},
	}
}

func (v *visitor) checkDocComment(fdec *ast.FuncDecl) {
	dtext := fdec.Doc.Text()
	lines := strings.Split(dtext, "\n")
	for i := range lines {
		if strings.Contains(lines[i], "DEPRECATED") {
			v.deprecated[fdec.Name.Name] = dtext
			logger.Printf("marked deprecated: %s\n", fdec.Name.Name)
		}
		if strings.Contains(lines[i], "PREVIEW") {
			v.preview[fdec.Name.Name] = dtext
			logger.Printf("marked preview: %s\n", fdec.Name.Name)
		}

		if lines[i] == "Implements:" {
			cfunc := cfuncFromComment(lines[i+1])
			if cfunc == "" {
				return
			}
			v.docMap[cfunc] = fdec.Name.Name
			logger.Printf("updated %s in doc map\n", cfunc)
		}
	}
}

func (v *visitor) checkCalled(s *ast.SelectorExpr) {
	ident, ok := s.X.(*ast.Ident)
	if !ok {
		return
	}
	if "C" == ident.String() {
		v.callMap[s.Sel.String()] = v.inFunction.Name.Name
		logger.Printf("updated %s in call map\n", s.Sel.String())
	}
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch {
	case node == nil:
		return nil
	case v.inFunction == nil:
	case node.Pos() > v.inFunction.End():
		logger.Printf("left function %v\n", v.inFunction.Name.Name)
		v.inFunction = nil
	}

	switch n := node.(type) {
	case *ast.File:
		v.inFunction = nil
		return v
	case *ast.FuncDecl:
		logger.Printf("checking function: %v\n", n.Name.Name)
		v.checkDocComment(n)
		v.inFunction = n
		return v
	case *ast.CallExpr:
		if v.inFunction == nil {
			return nil
		}
		if s, ok := n.Fun.(*ast.SelectorExpr); ok {
			v.checkCalled(s)
		}
	}
	if v.inFunction != nil {
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
