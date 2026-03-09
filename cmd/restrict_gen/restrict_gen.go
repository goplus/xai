/*
 * Copyright (c) 2026 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"os"
	"strconv"
	"strings"

	"github.com/qiniu/x/stringutil"
	"golang.org/x/tools/go/packages"
)

// -----------------------------------------------------------------------------

type fieldRestriction struct {
	orgName      string     // original field name
	newName      string     // new field name after rewrite
	typ          types.Type // field type
	enumSuggName string     // suggested enum type name, or empty
	doc          []string   // field doc comment, split by "." and trimmed
	stringEnum   []string   // string enum values, or nil
	docEnumVals  []string   // enum values parsed from doc comment, or nil
	notAllowedIf []string   // not allowed if the fields are provided, or nil
	optionalIf   []string   // optional if the fields are provided, or nil
	required     bool       // whether the field is required
}

func (p *fieldRestriction) hasRestriction() bool {
	return len(p.stringEnum)+len(p.optionalIf)+len(p.notAllowedIf) > 0 || p.required
}

type typeRestriction struct {
	typ    *types.Named
	fields []*fieldRestriction // restricted fields
}

func (p *typeRestriction) hasRestriction() bool {
	return len(p.fields) > 0
}

type pkgRestriction struct {
	pkgName string
	pkgPath string
	types   []*typeRestriction
}

func (p *pkgRestriction) hasRestriction() bool {
	return len(p.types) > 0
}

// -----------------------------------------------------------------------------

func decl(pkg *packages.Package, pos token.Pos) ast.Decl {
	for _, file := range pkg.Syntax {
		if file.Pos() <= pos && pos < file.End() {
			for _, decl := range file.Decls {
				if decl.Pos() <= pos && pos < decl.End() {
					return decl
				}
			}
		}
	}
	panic("decl not found")
}

func stringLit(v ast.Expr) string {
	if lit, ok := v.(*ast.BasicLit); ok {
		if lit.Kind == token.STRING {
			if val, err := strconv.Unquote(lit.Value); err == nil {
				return val
			}
		}
	}
	panic("stringLit failed")
}

func fieldDoc(def *ast.TypeSpec, name string) []string {
	if st, ok := def.Type.(*ast.StructType); ok {
		for _, field := range st.Fields.List {
			for _, fieldName := range field.Names {
				if fieldName.Name == name {
					doc := strings.ReplaceAll(field.Doc.Text(), "\n", " ")
					parts := strings.Split(doc, ". ")
					ret := parts[:0]
					for i, part := range parts {
						part = strings.TrimSpace(part)
						if part != "" {
							ret = append(ret, part)
							if i > 0 {
								log("  //", part)
							} else {
								log("\n  //", name+":", part)
							}
						}
					}
					return ret
				}
			}
		}
	}
	return nil
}

func required(doc []string) bool {
	return len(doc) == 0 || doc[0] != "Optional"
}

func getRewriteFlds(pkg *packages.Package) map[string]string {
	scope := pkg.Types.Scope()
	if o := scope.Lookup("rewriteFlds"); o != nil {
		def := decl(pkg, o.Pos()).(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
		val := def.Values[0].(*ast.CompositeLit)
		flds := make(map[string]string, len(val.Elts))
		for _, elt := range val.Elts {
			kv := elt.(*ast.KeyValueExpr)
			key := stringLit(kv.Key)
			value := stringLit(kv.Value)
			flds[key] = value
		}
		return flds
	}
	return nil
}

func collect(pkg *packages.Package) *pkgRestriction {
	pkgPath := pkg.PkgPath
	rewriteFlds := getRewriteFlds(pkg)
	log("package", pkgPath, rewriteFlds)
	scope := pkg.Types.Scope()
	names := scope.Names()
	ret := &pkgRestriction{pkgName: pkg.Name, pkgPath: pkgPath}
	for _, name := range names {
		o := scope.Lookup(name)
		if t, ok := o.Type().(*types.Named); ok {
			for i, n := 0, t.NumMethods(); i < n; i++ {
				mthd := t.Method(i)
				switch mthd.Name() {
				case "InputSchema":
					collectType(ret, pkg, t, rewriteFlds)
				}
			}
		}
	}
	return ret
}

func collectType(ret *pkgRestriction, pkg *packages.Package, t *types.Named, rewriteFlds map[string]string) {
	name := t.Obj().Name()
	log("\n==>", name)
	typ := &typeRestriction{typ: t}
	collectFields(typ, pkg, t, rewriteFlds)
	if typ.hasRestriction() {
		ret.types = append(ret.types, typ)
	}
}

func collectFields(ret *typeRestriction, pkg *packages.Package, t types.Type, rewriteFlds map[string]string) {
	typn, ok := t.(*types.Named)
	if !ok {
		return
	}
	if struc, ok := typn.Underlying().(*types.Struct); ok {
		o := typn.Obj()
		at := o.Pkg()
		if at != pkg.Types && at != nil {
			pkg = pkg.Imports[at.Path()]
		}
		def := decl(pkg, o.Pos()).(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
		for i, n := 0, struc.NumFields(); i < n; i++ {
			field := struc.Field(i)
			if field.Embedded() {
				collectFields(ret, pkg, field.Type(), rewriteFlds)
			} else if field.Exported() {
				orgName := field.Name()
				name := orgName
				if newName, ok := rewriteFlds[name]; ok {
					if newName == "" {
						continue
					}
					name = newName
				}
				typ := field.Type()
				if false && skipType(typ) {
					continue
				}
				doc := fieldDoc(def, orgName)
				field := &fieldRestriction{
					newName: name, orgName: orgName,
					typ: typ, doc: doc, required: required(doc),
				}
				collectRestrictionByDoc(field, doc)
				if tn, ok := typ.(*types.Named); ok {
					collectStringEnum(field, tn)
				} else if tb, ok := typ.(*types.Basic); ok && tb.Kind() == types.String {
					field.enumSuggName = "enum_" + at.Name() + "_" + name
					field.stringEnum = field.docEnumVals
					field.docEnumVals = nil
				}
				if field.hasRestriction() {
					ret.fields = append(ret.fields, field)
				}
			}
		}
	}
}

func checkOptionalIf(ret *fieldRestriction, part string) (ok bool) {
	const (
		prefix = "Optional if "
		suffix = " is provided"
	)
	left, ok := strings.CutPrefix(part, prefix)
	if ok {
		if text, has := strings.CutSuffix(left, suffix); has {
			vals := docValues(text, "or ", " or ")
			ret.optionalIf = append(ret.optionalIf, vals...)
		}
	}
	return
}

func checkNotAllowedIf(ret *fieldRestriction, part string) (ok bool) {
	const (
		prefix = "Not allowed if "
		suffix = " is provided"
	)
	left, ok := strings.CutPrefix(part, prefix)
	if ok {
		if text, has := strings.CutSuffix(left, suffix); has {
			vals := docValues(text, "or ", " or ")
			ret.notAllowedIf = append(ret.notAllowedIf, vals...)
		}
	}
	return
}

func checkEnumDoc(part string) (val string, ok bool) {
	const (
		prefix1 = "Supported values are: "
		prefix2 = "Supported values are "
		suffix  = " are supported"
	)
	val, ok = strings.CutSuffix(part, suffix)
	if !ok {
		val, ok = strings.CutPrefix(part, prefix1)
		if !ok {
			val, ok = strings.CutPrefix(part, prefix2)
		}
	}
	return
}

func docValues(text, cutPrefix, cutSep string) []string {
	vals := strings.Split(text, ", ")
	last := len(vals) - 1
	lastVal := vals[last]
	if v, ok := strings.CutPrefix(lastVal, cutPrefix); ok {
		vals[last] = v
	} else if v1, v2, ok := strings.Cut(lastVal, cutSep); ok {
		vals = vals[:last]
		vals = append(vals, v1, v2)
	}
	return vals
}

func collectRestrictionByDoc(ret *fieldRestriction, doc []string) {
	for _, part := range doc {
		if enumDoc, ok := checkEnumDoc(part); ok {
			log(" ", ret.newName, "DocEnumVals")
			docEnumVals := docValues(enumDoc, "and ", " and ")
			for i, val := range docEnumVals {
				n := len(val)
				if n > 2 && val[0] == '"' && val[n-1] == '"' {
					val = val[1 : n-1]
					docEnumVals[i] = val
				}
				log("   ", val)
			}
			ret.docEnumVals = docEnumVals
		} else if !checkOptionalIf(ret, part) {
			checkNotAllowedIf(ret, part)
		}
	}
	capitalizeNames(ret.newName, ret.optionalIf, "OptionalIf")
	capitalizeNames(ret.newName, ret.notAllowedIf, "NotAllowedIf")
}

func capitalizeNames(newName string, names []string, what string) {
	if len(names) > 0 {
		log(" ", newName, what)
		for i, name := range names {
			name = stringutil.Capitalize(name)
			names[i] = name
			log("   ", name)
		}
	}
}

func collectStringEnum(ret *fieldRestriction, tn *types.Named) {
	if tb, ok := tn.Underlying().(*types.Basic); ok && tb.Kind() == types.String {
		log(" ", ret.newName, tn)
		tobj := tn.Obj()
		scope := tobj.Pkg().Scope()
		names := scope.Names()
		for _, name := range names {
			obj := scope.Lookup(name)
			if c, ok := obj.(*types.Const); ok {
				if c.Type() == tn {
					val := constant.StringVal(c.Val())
					ret.stringEnum = append(ret.stringEnum, val)
					log("   ", val)
				}
			}
		}
		if len(ret.stringEnum) > 0 {
			ret.enumSuggName = "enum_" + suggestName(tobj)
		}
	}
}

func suggestName(o types.Object) string {
	return o.Pkg().Name() + "_" + o.Name()
}

func skipType(t types.Type) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	_, ok := t.(*types.Basic)
	return ok
}

// -----------------------------------------------------------------------------

func main() {
	fset := token.NewFileSet()
	conf := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedImports | packages.NeedDeps | packages.NeedSyntax,
		Fset: fset,
	}
	pkgs, _ := packages.Load(conf, ".")
	for _, pkg := range pkgs {
		ret := collect(pkg)
		if ret.hasRestriction() {
			gen(ret)
		}
	}
}

// -----------------------------------------------------------------------------

func log(v ...any) {
	fmt.Fprintln(os.Stderr, v...)
}

// -----------------------------------------------------------------------------
