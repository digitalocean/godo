// Generator: reads a vendor Go package's source (never executes it) and emits
// declarations that mirror its exported API, with stub bodies that return
// *unsupported.NotImplementedError.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var predeclared = map[string]bool{}

func init() {
	for _, n := range strings.Fields("bool byte complex64 complex128 error float32 float64 int int8 int16 int32 int64 rune string uint uint8 uint16 uint32 uint64 uintptr any comparable true false iota nil") {
		predeclared[n] = true
	}
}

type gen struct {
	fset       *token.FileSet
	files      []*ast.File
	pkgName    string
	product    string
	vendorPath string
	version    string
	hw         map[string]bool
	impl       map[string]bool
	stubIfaces map[string]bool
	importMap  map[string]string // qualifier -> import path (rewritten)
	closureQ   map[string]bool   // qualifiers whose used idents are recorded
	srcImports map[string]string // qualifier -> vendor import path
	types      map[string]*ast.TypeSpec
	typeOrder  []string
	ok         map[string]bool // exported types that survive
	report     []string
	used       map[string]map[string]bool
	needsRules []rule
	allow      map[string]bool
}

type rule struct {
	re    *regexp.Regexp
	needs string
}

func main() {
	var (
		src      = flag.String("src", "", "comma separated source dirs")
		pkg      = flag.String("pkg", "", "package name")
		out      = flag.String("out", "", "output file")
		product  = flag.String("product", "", "product name used in messages")
		vendor   = flag.String("vendor", "", "vendor import path (for header)")
		version  = flag.String("version", "", "vendor version (for header)")
		hw       = flag.String("hwtypes", "", "types declared by hand")
		impl     = flag.String("impl", "", "Type.Method or Func implemented by hand")
		stubIf   = flag.String("stubifaces", "", "interfaces to give a not-implemented implementation")
		imap     = flag.String("importmap", "", "qualifier=path,...")
		closureQ = flag.String("closureq", "", "qualifiers whose used identifiers are recorded")
		usedOut  = flag.String("usedout", "", "write used closure identifiers (json) here")
		mode     = flag.String("mode", "stubs", "stubs | closure | copy")
		roots    = flag.String("roots", "", "closure roots")
		rulesF   = flag.String("rules", "", "needs rules file: regexp<TAB>text per line")
		drop     = flag.String("drop", "", "copy mode: funcs to drop")
		allow    = flag.String("allow", "", "unexported identifiers that are declared by hand")
		verbatim = flag.String("verbatim", "", "source file names copied as they are, not stubbed")
		verbOut  = flag.String("verbatimout", "", "where the verbatim files go")
	)
	flag.Parse()
	g := &gen{
		fset: token.NewFileSet(), pkgName: *pkg, product: *product, vendorPath: *vendor, version: *version,
		hw: set(*hw), allow: set(*allow), impl: set(*impl), stubIfaces: set(*stubIf), importMap: map[string]string{}, closureQ: set(*closureQ),
		srcImports: map[string]string{}, types: map[string]*ast.TypeSpec{}, ok: map[string]bool{}, used: map[string]map[string]bool{},
	}
	for _, kv := range split(*imap) {
		p := strings.SplitN(kv, "=", 2)
		g.importMap[p[0]] = p[1]
	}
	if *rulesF != "" {
		b, err := os.ReadFile(*rulesF)
		check(err)
		for _, ln := range strings.Split(string(b), "\n") {
			if strings.TrimSpace(ln) == "" || strings.HasPrefix(ln, "#") {
				continue
			}
			p := strings.SplitN(ln, "\t", 2)
			g.needsRules = append(g.needsRules, rule{regexp.MustCompile(p[0]), p[1]})
		}
	}
	var vfiles []*ast.File
	for _, d := range split(*src) {
		ents, err := os.ReadDir(d)
		check(err)
		for _, e := range ents {
			n := e.Name()
			if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
				continue
			}
			f, err := parser.ParseFile(g.fset, filepath.Join(d, n), nil, parser.ParseComments)
			check(err)
			if set(*verbatim)[n] {
				vfiles = append(vfiles, f)
				continue
			}
			g.files = append(g.files, f)
		}
	}
	for _, f := range g.files {
		for _, im := range f.Imports {
			path := strings.Trim(im.Path.Value, `"`)
			name := filepath.Base(path)
			if im.Name != nil {
				name = im.Name.Name
			}
			g.srcImports[name] = path
		}
	}
	if len(vfiles) > 0 {
		g2 := &gen{fset: g.fset, files: vfiles, pkgName: g.pkgName, vendorPath: g.vendorPath, version: g.version, importMap: g.importMap}
		b := g2.copyMode(nil)
		check(os.WriteFile(*verbOut, b.Bytes(), 0o644))
	}
	var buf bytes.Buffer
	switch *mode {
	case "stubs":
		buf = g.stubs()
	case "closure":
		buf = g.closure(split(*roots))
	case "copy":
		buf = g.copyMode(set(*drop))
	}
	check(os.WriteFile(*out, buf.Bytes(), 0o644))
	if *usedOut != "" {
		m := map[string][]string{}
		for q, s := range g.used {
			for id := range s {
				m[q] = append(m[q], id)
			}
			sort.Strings(m[q])
		}
		b, _ := json.MarshalIndent(m, "", " ")
		check(os.WriteFile(*usedOut, b, 0o644))
	}
	for _, r := range g.report {
		fmt.Fprintln(os.Stderr, "report:", r)
	}
}

func set(s string) map[string]bool {
	m := map[string]bool{}
	for _, x := range split(s) {
		m[x] = true
	}
	return m
}
func split(s string) []string {
	var o []string
	for _, x := range strings.Split(s, ",") {
		if x = strings.TrimSpace(x); x != "" {
			o = append(o, x)
		}
	}
	return o
}
func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (g *gen) pr(n any) string {
	var b bytes.Buffer
	check(printer.Fprint(&b, g.fset, n))
	return b.String()
}

func isStd(path string) bool { return !strings.Contains(strings.Split(path, "/")[0], ".") }

// qualifierOK reports whether pkg qualifier q may appear in emitted code.
func (g *gen) qualifierOK(q string) bool {
	if _, ok := g.importMap[q]; ok {
		return true
	}
	p, ok := g.srcImports[q]
	return ok && isStd(p)
}

// mention returns "" when expr only references allowed identifiers, else why not.
func (g *gen) mention(expr ast.Node, extra map[string]bool) string {
	reason := ""
	var walk func(n ast.Node) bool
	walk = func(n ast.Node) bool {
		if reason != "" {
			return false
		}
		switch x := n.(type) {
		case *ast.SelectorExpr:
			if id, ok := x.X.(*ast.Ident); ok && g.isQualifier(id.Name) {
				if !g.qualifierOK(id.Name) {
					reason = "package " + id.Name
				}
				return false
			}
			return true
		case *ast.Ident:
			if predeclared[x.Name] {
				return false
			}
			if ast.IsExported(x.Name) {
				if _, ok := g.types[x.Name]; ok {
					if !g.ok[x.Name] && g.ok != nil && len(g.ok) > 0 && !g.hw[x.Name] {
						reason = "type " + x.Name
					}
					return false
				}
				return false // exported const/var/func identifier: allow
			}
			if g.allow[x.Name] || (extra != nil && extra[x.Name]) {
				return false
			}
			reason = "unexported " + x.Name
			return false
		case *ast.FuncType:
			// walk only types of params/results, not names
			for _, fl := range []*ast.FieldList{x.TypeParams, x.Params, x.Results} {
				if fl == nil {
					continue
				}
				for _, f := range fl.List {
					ast.Inspect(f.Type, walk)
				}
			}
			return false
		case *ast.StructType, *ast.InterfaceType:
			for _, f := range fieldsOf(n) {
				ast.Inspect(f.Type, walk)
			}
			return false
		}
		return true
	}
	ast.Inspect(expr, walk)
	return reason
}

func (g *gen) isQualifier(name string) bool {
	_, ok := g.srcImports[name]
	return ok
}

func fieldsOf(n ast.Node) []*ast.Field {
	switch x := n.(type) {
	case *ast.StructType:
		return x.Fields.List
	case *ast.InterfaceType:
		return x.Methods.List
	}
	return nil
}

func (g *gen) collectTypes() {
	for _, f := range g.files {
		for _, d := range f.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, s := range gd.Specs {
				ts := s.(*ast.TypeSpec)
				if _, dup := g.types[ts.Name.Name]; !dup {
					g.types[ts.Name.Name] = ts
					g.typeOrder = append(g.typeOrder, ts.Name.Name)
				}
			}
		}
	}
}

// strippedType returns the type expression with unexported / unusable struct fields removed.
func (g *gen) strippedType(ts *ast.TypeSpec) (ast.Expr, string) {
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		if it, ok := ts.Type.(*ast.InterfaceType); ok {
			for _, f := range it.Methods.List {
				if r := g.mention(f.Type, nil); r != "" {
					return nil, r
				}
			}
			return ts.Type, ""
		}
		if r := g.mention(ts.Type, nil); r != "" {
			return nil, r
		}
		return ts.Type, ""
	}
	nl := &ast.FieldList{}
	for _, f := range st.Fields.List {
		if g.mention(f.Type, nil) != "" {
			continue
		}
		if len(f.Names) == 0 {
			if id := baseIdent(f.Type); id != "" && ast.IsExported(id) {
				nl.List = append(nl.List, f)
			}
			continue
		}
		var names []*ast.Ident
		for _, n := range f.Names {
			if ast.IsExported(n.Name) {
				names = append(names, n)
			}
		}
		if len(names) > 0 {
			nf := *f
			nf.Names = names
			nf.Doc, nf.Comment = nil, nil
			nl.List = append(nl.List, &nf)
		}
	}
	return &ast.StructType{Struct: st.Struct, Fields: nl}, ""
}

func baseIdent(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return baseIdent(x.X)
	case *ast.SelectorExpr:
		return x.Sel.Name
	}
	return ""
}

func (g *gen) decideTypes() {
	g.collectTypes()
	for _, n := range g.typeOrder {
		if ast.IsExported(n) {
			g.ok[n] = true
		}
	}
	for changed := true; changed; {
		changed = false
		for _, n := range g.typeOrder {
			if !g.ok[n] || g.hw[n] {
				continue
			}
			if _, r := g.strippedType(g.types[n]); r != "" {
				g.ok[n] = false
				g.report = append(g.report, "drop type "+n+": "+r)
				changed = true
			}
		}
	}
}

func (g *gen) noteUsed(n ast.Node) map[string]bool {
	qs := map[string]bool{}
	ast.Inspect(n, func(x ast.Node) bool {
		if se, ok := x.(*ast.SelectorExpr); ok {
			if id, ok := se.X.(*ast.Ident); ok && g.isQualifier(id.Name) {
				qs[id.Name] = true
				if g.closureQ[id.Name] {
					if g.used[id.Name] == nil {
						g.used[id.Name] = map[string]bool{}
					}
					g.used[id.Name][se.Sel.Name] = true
				}
			}
		}
		return true
	})
	return qs
}

func (g *gen) needs(typ, method string) string {
	key := typ + "." + method
	for _, r := range g.needsRules {
		if r.re.MatchString(key) {
			return r.needs
		}
	}
	return "a matching DigitalOcean Sandboxes API"
}

func (g *gen) stubBody(recvName string, recvType string, name string, ft *ast.FuncType, typ string) string {
	label := name
	if typ != "" {
		label = typ + "." + name
	}
	call := fmt.Sprintf("unsupported.New(%q, %q, %q)", g.product, label, g.needs(typ, name))
	var res []string
	if ft.Results != nil {
		for _, f := range ft.Results.List {
			n := len(f.Names)
			if n == 0 {
				n = 1
			}
			for i := 0; i < n; i++ {
				res = append(res, g.pr(f.Type))
			}
		}
	}
	if len(res) == 0 {
		return "\tpanic(" + call + ")\n"
	}
	if res[len(res)-1] == "error" {
		var zs []string
		for _, r := range res[:len(res)-1] {
			zs = append(zs, "*new("+r+")")
		}
		zs = append(zs, call)
		return "\treturn " + strings.Join(zs, ", ") + "\n"
	}
	if len(res) == 1 {
		r := res[0]
		if m := regexp.MustCompile(`^iter\.Seq2\[(.+), error\]$`).FindStringSubmatch(r); m != nil {
			return fmt.Sprintf("\treturn func(yield func(%s, error) bool) { yield(*new(%s), %s) }\n", m[1], m[1], call)
		}
		if m := regexp.MustCompile(`^\*([A-Z]\w*Service)$`).FindStringSubmatch(r); m != nil && strings.HasPrefix(name, "Create") {
			return "\treturn &" + m[1] + "{}\n"
		}
		if m := regexp.MustCompile(`^\*([A-Z]\w*)$`).FindStringSubmatch(r); m != nil && typ == "" {
			if ts, ok := g.types[m[1]]; ok && g.ok[m[1]] {
				if _, isStruct := ts.Type.(*ast.StructType); isStruct {
					return "\treturn &" + m[1] + "{}\n"
				}
			}
		}
		if recvName != "" && r == recvType && strings.HasPrefix(r, "*") {
			return "\treturn " + recvName + "\n"
		}
	}
	return "\tpanic(" + call + ")\n"
}

func recvInfo(fd *ast.FuncDecl) (name, typeText, base string) {
	f := fd.Recv.List[0]
	if len(f.Names) > 0 {
		name = f.Names[0].Name
	}
	t := f.Type
	star := ""
	if s, ok := t.(*ast.StarExpr); ok {
		star = "*"
		t = s.X
	}
	if id, ok := t.(*ast.Ident); ok {
		base = id.Name
	}
	return name, star + base, base
}

func (g *gen) sigText(ft *ast.FuncType) string {
	s := g.pr(ft)
	return strings.TrimPrefix(s, "func")
}

func (g *gen) stubs() bytes.Buffer {
	g.decideTypes()
	var body bytes.Buffer
	quals := map[string]bool{}
	stubCount := 0
	emit := func(s string, nodes ...ast.Node) {
		body.WriteString(s)
		for _, n := range nodes {
			for q := range g.noteUsed(n) {
				quals[q] = true
			}
		}
	}
	// types
	for _, n := range g.typeOrder {
		if !g.ok[n] || g.hw[n] || !ast.IsExported(n) {
			continue
		}
		ts := g.types[n]
		typ, _ := g.strippedType(ts)
		ns := &ast.TypeSpec{Name: ts.Name, Assign: ts.Assign, Type: typ}
		gd := &ast.GenDecl{Tok: token.TYPE, Specs: []ast.Spec{ns}}
		emit(fmt.Sprintf("// %s mirrors %s.%s.\n%s\n\n", n, g.vendorPath, n, g.pr(gd)), typ)
	}
	// consts and vars
	for _, f := range g.files {
		for _, d := range f.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || (gd.Tok != token.CONST && gd.Tok != token.VAR) {
				continue
			}
			var keep []ast.Spec
			complete := true
			for _, s := range gd.Specs {
				vs := s.(*ast.ValueSpec)
				good := true
				for _, n := range vs.Names {
					if !ast.IsExported(n.Name) {
						good = false
					}
				}
				if good && vs.Type != nil && g.mention(vs.Type, nil) != "" {
					good = false
				}
				for _, v := range vs.Values {
					if g.mention(v, nil) != "" {
						good = false
					}
				}
				if good && gd.Tok == token.CONST && len(vs.Values) == 0 && vs.Type == nil {
					// implicit iota repetition
					good = true
				}
				if good {
					cp := *vs
					cp.Doc, cp.Comment = nil, nil
					keep = append(keep, &cp)
				} else {
					complete = false
				}
			}
			if len(keep) == 0 {
				continue
			}
			if !complete {
				for _, s := range keep {
					if vs := s.(*ast.ValueSpec); len(vs.Values) == 0 && gd.Tok == token.CONST {
						g.report = append(g.report, "const block partially kept (implicit values): "+vs.Names[0].Name)
					}
				}
			}
			ng := &ast.GenDecl{Tok: gd.Tok, Specs: keep}
			if len(keep) > 1 {
				ng.Lparen = 1
			}
			emit(g.pr(ng)+"\n\n", ng)
		}
	}
	// funcs and methods
	for _, f := range g.files {
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || !ast.IsExported(fd.Name.Name) || fd.Type.TypeParams != nil {
				continue
			}
			if fd.Recv == nil {
				if g.impl[fd.Name.Name] {
					continue
				}
				if r := g.mention(fd.Type, nil); r != "" {
					g.report = append(g.report, "skip func "+fd.Name.Name+": "+r)
					continue
				}
				emit(fmt.Sprintf("// %s mirrors %s.%s.\nfunc %s%s {\n%s}\n\n", fd.Name.Name, g.vendorPath, fd.Name.Name, fd.Name.Name, g.sigText(fd.Type), g.stubBody("", "", fd.Name.Name, fd.Type, "")), fd.Type)
				stubCount++
				continue
			}
			rn, rt, base := recvInfo(fd)
			if !ast.IsExported(base) || !g.ok[base] {
				continue
			}
			if g.impl[base+"."+fd.Name.Name] {
				continue
			}
			if r := g.mention(fd.Type, nil); r != "" {
				g.report = append(g.report, "skip method "+base+"."+fd.Name.Name+": "+r)
				continue
			}
			recv := rt
			if rn != "" {
				recv = rn + " " + rt
			}
			emit(fmt.Sprintf("// %s mirrors %s.%s.%s.\nfunc (%s) %s%s {\n%s}\n\n", fd.Name.Name, g.vendorPath, base, fd.Name.Name, recv, fd.Name.Name, g.sigText(fd.Type), g.stubBody(rn, rt, fd.Name.Name, fd.Type, base)), fd.Type)
			stubCount++
		}
	}
	// interface stubs
	for _, n := range g.typeOrder {
		if !g.stubIfaces[n] {
			continue
		}
		it, ok := g.types[n].Type.(*ast.InterfaceType)
		if !ok {
			continue
		}
		sn := "notImplemented" + n
		emit(fmt.Sprintf("// %s implements %s by returning *unsupported.NotImplementedError from every method.\ntype %s struct{}\n\n", sn, n, sn))
		for _, m := range it.Methods.List {
			if len(m.Names) == 0 {
				continue
			}
			ft := m.Type.(*ast.FuncType)
			if g.impl[n+"."+m.Names[0].Name] {
				continue
			}
			emit(fmt.Sprintf("func (%s) %s%s {\n%s}\n\n", sn, m.Names[0].Name, g.sigText(ft), g.stubBody("", "", m.Names[0].Name, ft, n)), ft)
			stubCount++
		}
	}
	var out bytes.Buffer
	out.WriteString("// Code generated by sandbox-compat-gen; DO NOT EDIT.\n//\n")
	fmt.Fprintf(&out, "// Declarations mirror %s %s (Apache-2.0); bodies are DigitalOcean stubs.\n\n", g.vendorPath, g.version)
	fmt.Fprintf(&out, "package %s\n\nimport (\n", g.pkgName)
	var std, other []string
	for q := range quals {
		if p, ok := g.importMap[q]; ok {
			other = append(other, fmt.Sprintf("\t%s %q", q, p))
		} else if p, ok := g.srcImports[q]; ok && isStd(p) {
			std = append(std, fmt.Sprintf("\t%q", p))
		}
	}
	if stubCount > 0 {
		other = append(other, "\tunsupported \"github.com/digitalocean/godo/sandbox/compat/unsupported\"")
	}
	sort.Strings(std)
	sort.Strings(other)
	for _, s := range std {
		out.WriteString(s + "\n")
	}
	out.WriteString("\n")
	for _, s := range other {
		out.WriteString(s + "\n")
	}
	out.WriteString(")\n\n")
	out.Write(body.Bytes())
	return out
}

// closure emits the named types and everything they reference in the same package, plus typed consts.
func (g *gen) closure(roots []string) bytes.Buffer {
	g.collectTypes()
	in := map[string]bool{}
	var order []string
	var visit func(n string)
	visit = func(n string) {
		if in[n] {
			return
		}
		ts, ok := g.types[n]
		if !ok {
			g.report = append(g.report, "root not found: "+n)
			return
		}
		in[n] = true
		order = append(order, n)
		ast.Inspect(ts.Type, func(x ast.Node) bool {
			if id, ok := x.(*ast.Ident); ok {
				if _, isT := g.types[id.Name]; isT && id.Name != n {
					visit(id.Name)
				}
			}
			return true
		})
	}
	for _, r := range roots {
		visit(r)
	}
	sort.Strings(order)
	var body bytes.Buffer
	quals := map[string]bool{}
	for _, n := range order {
		ts := g.types[n]
		typ := ts.Type
		if st, ok := typ.(*ast.StructType); ok {
			nl := &ast.FieldList{}
			for _, f := range st.Fields.List {
				if len(f.Names) == 0 {
					continue
				}
				var names []*ast.Ident
				for _, nm := range f.Names {
					if ast.IsExported(nm.Name) {
						names = append(names, nm)
					}
				}
				if len(names) > 0 {
					nf := *f
					nf.Names = names
					nf.Doc, nf.Comment = nil, nil
					nl.List = append(nl.List, &nf)
				}
			}
			typ = &ast.StructType{Struct: st.Struct, Fields: nl}
		}
		ns := &ast.TypeSpec{Name: ts.Name, Assign: ts.Assign, Type: typ}
		gd := &ast.GenDecl{Tok: token.TYPE, Specs: []ast.Spec{ns}}
		body.WriteString(fmt.Sprintf("// %s mirrors %s.%s (declaration only).\n%s\n\n", n, g.vendorPath, n, g.pr(gd)))
		for q := range g.noteUsed(typ) {
			quals[q] = true
		}
	}
	// typed consts
	for _, f := range g.files {
		for _, d := range f.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			var keep []ast.Spec
			for _, s := range gd.Specs {
				vs := s.(*ast.ValueSpec)
				if id, ok := vs.Type.(*ast.Ident); ok && in[id.Name] {
					cp := *vs
					cp.Doc, cp.Comment = nil, nil
					keep = append(keep, &cp)
				}
			}
			if len(keep) == 0 {
				continue
			}
			ng := &ast.GenDecl{Tok: token.CONST, Specs: keep}
			if len(keep) > 1 {
				ng.Lparen = 1
			}
			body.WriteString(g.pr(ng) + "\n\n")
		}
	}
	var out bytes.Buffer
	out.WriteString("// Code generated by sandbox-compat-gen; DO NOT EDIT.\n//\n")
	fmt.Fprintf(&out, "// Type declarations mirror %s %s (Apache-2.0). Methods are not mirrored.\n\n", g.vendorPath, g.version)
	fmt.Fprintf(&out, "package %s\n\n", g.pkgName)
	var std []string
	for q := range quals {
		if p, ok := g.srcImports[q]; ok && isStd(p) {
			std = append(std, fmt.Sprintf("\t%q", p))
		}
	}
	sort.Strings(std)
	if len(std) > 0 {
		out.WriteString("import (\n" + strings.Join(std, "\n") + "\n)\n\n")
	}
	out.Write(body.Bytes())
	return out
}

// copyMode prints each source file again with import paths rewritten, dropping named funcs.
func (g *gen) copyMode(drop map[string]bool) bytes.Buffer {
	var out bytes.Buffer
	out.WriteString("// Copied from " + g.vendorPath + " " + g.version + " (Apache-2.0); see NOTICE.\n\n")
	fmt.Fprintf(&out, "package %s\n\n", g.pkgName)
	imports := map[string]string{}
	var bodies bytes.Buffer
	for _, f := range g.files {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil && drop[fd.Name.Name] {
				continue
			}
			if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
				continue
			}
			cn := &printer.CommentedNode{Node: d, Comments: commentsFor(f, d)}
			check((&printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}).Fprint(&bodies, g.fset, cn))
			bodies.WriteString("\n\n")
		}
		for _, im := range f.Imports {
			path := strings.Trim(im.Path.Value, `"`)
			name := filepath.Base(path)
			if im.Name != nil {
				name = im.Name.Name
			}
			imports[name] = path
		}
	}
	text := bodies.String()
	var lines []string
	for name, path := range imports {
		if np, ok := g.importMap[name]; ok {
			path = np
		}
		if !regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\.`).MatchString(text) {
			continue
		}
		lines = append(lines, fmt.Sprintf("\t%s %q", name, path))
	}
	sort.Strings(lines)
	if len(lines) > 0 {
		out.WriteString("import (\n" + strings.Join(lines, "\n") + "\n)\n\n")
	}
	out.WriteString(text)
	return out
}

func commentsFor(f *ast.File, d ast.Decl) []*ast.CommentGroup {
	var cs []*ast.CommentGroup
	for _, c := range f.Comments {
		if c.Pos() >= d.Pos()-1 && c.End() <= d.End() {
			cs = append(cs, c)
		}
	}
	return cs
}
