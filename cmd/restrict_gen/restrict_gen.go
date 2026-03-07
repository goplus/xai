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
	"go/constant"
	"go/token"
	"go/types"
	"os"

	"github.com/goplus/gogen"
	"golang.org/x/tools/go/packages"
)

// -----------------------------------------------------------------------------

var geminiRewriteFlds = map[string]string{
	"HTTPOptions":     "",
	"SDKHTTPResponse": "",
	"Labels":          "",
	"OutputGCSURI":    "OutputStgUri",
}

var pkgRewriteFlds = map[string]map[string]string{
	"github.com/goplus/xai/spec/gemini": geminiRewriteFlds,
}

// -----------------------------------------------------------------------------

type fieldRestriction struct {
	name       string   // field name
	stringEnum []string // string enum values, or nil
}

func (p *fieldRestriction) hasRestriction() bool {
	return len(p.stringEnum) > 0
}

type typeRestriction struct {
	typ    *types.Named
	fields []fieldRestriction // restricted fields
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

func fieldIndex(t types.Type, name string) int {
	if struc, ok := t.Underlying().(*types.Struct); ok {
		for i, n := 0, struc.NumFields(); i < n; i++ {
			field := struc.Field(i)
			if field.Name() == name {
				return i
			}
		}
	}
	panic("fieldIndex failed: " + name)
}

func gen(ret *pkgRestriction) {
	out := gogen.NewPackage(ret.pkgPath, ret.pkgName, nil)
	xai := out.Import("github.com/goplus/xai/spec")
	str := types.Typ[types.String]
	strSlice := types.NewSlice(str)
	restr := xai.Ref("Restriction").Type()
	stringEnum := xai.Ref("StringEnum").Type()
	iValues := fieldIndex(stringEnum, "Values")
	iLimit := fieldIndex(restr, "Limit")
	ptrRestr := types.NewPointer(restr)           // *xai.Restriction
	mapNameToRestr := types.NewMap(str, ptrRestr) // map[string]*xai.Restriction
	scope := out.Types.Scope()
	for _, r := range ret.types {
		typName := r.typ.Obj().Name()
		log("==> restriction", typName)
		name := "restriction_" + typName
		out.NewVarDefs(scope).NewAndInit(func(cb *gogen.CodeBuilder) int {
			flds := r.fields
			for _, fld := range flds {
				log("  ", fld.name, fld.stringEnum)
				cb.Val(fld.name)
				cb.Val(iLimit)
				if vals := fld.stringEnum; len(vals) > 0 {
					cb.Val(iValues)
					for _, val := range vals {
						cb.Val(val)
					}
					cb.SliceLit(strSlice, len(vals)).
						StructLit(stringEnum, 2, true).UnaryOp(token.AND)
				}
				cb.StructLit(restr, 2, true).UnaryOp(token.AND)
			}
			cb.MapLit(mapNameToRestr, len(flds)<<1)
			return 1
		}, token.NoPos, nil, name)
	}
	err := out.WriteTo(os.Stdout)
	if err != nil {
		log("genRestriction failed:", err)
		os.Exit(1)
	}
}

// -----------------------------------------------------------------------------

func collect(pkg *packages.Package) *pkgRestriction {
	pkgPath := pkg.PkgPath
	rewriteFlds := pkgRewriteFlds[pkgPath]
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
					collectType(ret, t, rewriteFlds)
				}
			}
		}
	}
	return ret
}

func collectType(ret *pkgRestriction, t *types.Named, rewriteFlds map[string]string) {
	name := t.Obj().Name()
	log("==>", name)
	typ := &typeRestriction{typ: t}
	collectFields(typ, t, rewriteFlds)
	if typ.hasRestriction() {
		ret.types = append(ret.types, typ)
	}
}

func collectFields(ret *typeRestriction, t types.Type, rewriteFlds map[string]string) {
	if struc, ok := t.Underlying().(*types.Struct); ok {
		for i, n := 0, struc.NumFields(); i < n; i++ {
			field := struc.Field(i)
			if field.Embedded() {
				collectFields(ret, field.Type(), rewriteFlds)
			} else if field.Exported() {
				name := field.Name()
				if newName, ok := rewriteFlds[name]; ok {
					if newName == "" {
						continue
					}
					name = newName
				}
				typ := field.Type()
				if skipType(typ) {
					continue
				}
				field := &fieldRestriction{name: name}
				if tn, ok := typ.(*types.Named); ok {
					collectStringEnum(field, name, tn)
				}
				if field.hasRestriction() {
					ret.fields = append(ret.fields, *field)
				}
			}
		}
	}
}

func collectStringEnum(ret *fieldRestriction, name string, tn *types.Named) {
	if tb, ok := tn.Underlying().(*types.Basic); ok && tb.Kind() == types.String {
		log(" ", name, tn)
		scope := tn.Obj().Pkg().Scope()
		names := scope.Names()
		for _, name := range names {
			o := scope.Lookup(name)
			if c, ok := o.(*types.Const); ok {
				if c.Type() == tn {
					val := constant.StringVal(c.Val())
					ret.stringEnum = append(ret.stringEnum, val)
					log("   ", val)
				}
			}
		}
	}
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
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
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
