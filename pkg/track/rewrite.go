package track

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

/**
@file:
@author: levi.Tang
@time: 2024/10/28 19:34
@description:
**/

const _defaultImport = "github.com/toheart/golearn/observability/functrace"

var debug = false

// RewriteDir
//
//	@Description: 对目录中所有文件进行重写
//	@param dir
func RewriteDir(dir string) {
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("path: %s, walk found err: %s \n ", path, err.Error())
			return nil
		}
		if info.IsDir() {
			return nil
		}
		// 排除
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fullPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Printf("path: %s, get abs filepath err: %s \n ", path, err.Error())
			return err
		}
		r, err := NewRewrite(fullPath)
		if err != nil {
			fmt.Printf("path: %s, get abs filepath err: %s \n ", path, err.Error())
			return err
		}
		r.RewriteFile()
		fmt.Printf("path: %s rewrite success \n", path)
		return nil
	})
	if err != nil {
		fmt.Printf("filepath walk found err:%s \n", err.Error())
	}
}

func NewRewrite(fullPath string) (*Rewrite, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fullPath, nil, 0)
	if err != nil {
		return nil, err
	}

	return &Rewrite{fullPath: fullPath,
		fset: fset,
		f:    f}, nil
}

type Rewrite struct {
	fullPath string
	fset     *token.FileSet
	f        *ast.File
}

func (r *Rewrite) genTraceParams(funcType *ast.FuncType) []ast.Expr {
	if len(funcType.Params.List) == 0 {
		return nil
	}
	var params []string
	for _, item := range funcType.Params.List {
		for _, j := range item.Names {
			params = append(params, j.Name)
		}
	}

	var elts []ast.Expr
	for _, param := range params {
		elts = append(elts, &ast.BasicLit{
			Kind:  token.VAR,
			Value: param,
		})
	}
	return elts
}

// genDefer
//
//	@Description: 用于生成defer函数
//	@receiver r
//	@param elts
//	@return *ast.DeferStmt
func (r *Rewrite) genDefer(elts []ast.Expr) *ast.DeferStmt {
	// 创建字符串切片参数
	sliceParams := &ast.CompositeLit{
		Type: &ast.ArrayType{
			Elt: &ast.InterfaceType{ // 空接口
				Methods: &ast.FieldList{
					List: []*ast.Field{},
				},
			},
		},
		Elts: elts,
	}
	// 创建函数调用表达式
	callExpr := &ast.CallExpr{
		Fun: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent("functrace"),
				Sel: ast.NewIdent("Trace"),
			},
			Args: []ast.Expr{sliceParams},
		},
	}
	return &ast.DeferStmt{
		Call: callExpr,
	}
}

func (r *Rewrite) ImportFunctrace() {
	importFlag := false
	for _, item := range r.f.Imports {
		if item.Path.Value == _defaultImport {
			importFlag = true
			break
		}
	}
	// 如果没有导入包, 则将functrace包导入
	if !importFlag {
		astutil.AddImport(r.fset, r.f, "github.com/toheart/golearn/observability/functrace")
	}
}

func (r *Rewrite) HasSameDefer(decl *ast.FuncDecl) bool {
	for _, stmt := range decl.Body.List {
		// 判断是否为defer 函数
		ds, ok := stmt.(*ast.DeferStmt)
		if !ok {
			continue
		}

		ce, ok := ds.Call.Fun.(*ast.CallExpr)
		if !ok {
			continue
		}
		se, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		x, ok := se.X.(*ast.Ident)
		if !ok {
			continue
		}
		if (x.Name == "functrace") && (se.Sel.Name == "Trace") {
			// 已经存在直接返回
			return true
		}
	}
	return false
}

func (r *Rewrite) RewriteFile() {
	// 插入import
	r.ImportFunctrace()
	// 插入defer函数
	for _, item := range r.f.Decls {
		funcDel, ok := item.(*ast.FuncDecl)
		if !ok {
			continue
		}
		// 判断是否需要插入defer函数
		if r.HasSameDefer(funcDel) {
			continue
		}
		elts := r.genTraceParams(funcDel.Type)
		deferStmt := r.genDefer(elts)
		// 将defer语句添加到函数体的开头
		funcDel.Body.List = append([]ast.Stmt{deferStmt}, funcDel.Body.List...)
	}
	buf := &bytes.Buffer{}
	err := format.Node(buf, r.fset, r.f)
	if err != nil {
		return
	}
	if debug {
		fmt.Println(buf.String())
		return
	}
	if err = os.WriteFile(r.fullPath, buf.Bytes(), 0666); err != nil {
		fmt.Printf("write %s error: %v\n", r.fullPath, err)
		return
	}
}
