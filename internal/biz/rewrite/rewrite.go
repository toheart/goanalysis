package rewrite

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

/**
@file:
@author: levi.Tang
@time: 2024/10/28 19:34
@description:
**/

const _defaultImport = "github.com/toheart/functrace"

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
			// 跳过vendor目录
			if info.Name() == "vendor" {
				return filepath.SkipDir
			}
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
	f, err := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return &Rewrite{
		fullPath: fullPath,
		fset:     fset,
		f:        f,
	}, nil
}

type Rewrite struct {
	fullPath string
	fset     *token.FileSet
	f        *ast.File
}

func (r *Rewrite) genTraceParams(funcType *ast.FuncType, recv *ast.FieldList) []ast.Expr {
	var params []string

	// 处理接收器
	if recv != nil && len(recv.List) > 0 {
		for _, field := range recv.List {
			for _, name := range field.Names {
				if name.Name != "_" {
					params = append(params, name.Name)
				}
			}
		}
	}

	// 处理函数参数
	if funcType.Params != nil && len(funcType.Params.List) > 0 {
		for _, item := range funcType.Params.List {
			for _, j := range item.Names {
				if j.Name == "_" {
					continue
				}
				params = append(params, j.Name)
			}
		}
	}

	if len(params) == 0 {
		return nil
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
			Args: []ast.Expr{
				sliceParams,
			},
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
		astutil.AddImport(r.fset, r.f, _defaultImport)
	}
}

func (r *Rewrite) HasSameDefer(decl *ast.FuncDecl) bool {
	// 检查函数是否有函数体
	if decl.Body == nil {
		return false
	}

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

// hasSameDeferInBody 通用的defer检查函数，用于检查任何函数体中是否已有相同的defer语句
func (r *Rewrite) hasSameDeferInBody(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}

	for _, stmt := range body.List {
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
			return true
		}
	}
	return false
}

// addDeferToBody 通用的添加defer语句函数
func (r *Rewrite) addDeferToBody(body *ast.BlockStmt, funcType *ast.FuncType, recv *ast.FieldList) bool {
	if body == nil {
		return false
	}

	// 检查是否已经有相同的defer语句
	if r.hasSameDeferInBody(body) {
		return false
	}

	// 生成defer语句
	elts := r.genTraceParams(funcType, recv)
	deferStmt := r.genDefer(elts)

	// 将defer语句添加到函数体的开头
	body.List = append([]ast.Stmt{deferStmt}, body.List...)
	return true
}

// processFuncLit 处理函数字面量，添加defer语句
func (r *Rewrite) processFuncLit(funcLit *ast.FuncLit) bool {
	return r.addDeferToBody(funcLit.Body, funcLit.Type, nil)
}

// processGoStmt 处理go语句
func (r *Rewrite) processGoStmt(goStmt *ast.GoStmt) bool {
	// 检查go语句中的函数是否为函数字面量
	if funcLit, ok := goStmt.Call.Fun.(*ast.FuncLit); ok {
		return r.processFuncLit(funcLit)
	}

	// 处理go语句中的函数调用参数，查找其中的函数字面量
	return r.processCallExpr(goStmt.Call)
}

// processCallExpr 处理函数调用表达式，查找其中的函数字面量参数
func (r *Rewrite) processCallExpr(callExpr *ast.CallExpr) bool {
	modified := false

	// 检查函数调用的参数中是否有函数字面量
	for _, arg := range callExpr.Args {
		if funcLit, ok := arg.(*ast.FuncLit); ok {
			if r.processFuncLit(funcLit) {
				modified = true
			}
		} else if nestedCall, ok := arg.(*ast.CallExpr); ok {
			// 递归处理嵌套的函数调用
			if r.processCallExpr(nestedCall) {
				modified = true
			}
		}
	}

	// 递归处理函数调用本身（如果它也是一个函数字面量）
	if funcLit, ok := callExpr.Fun.(*ast.FuncLit); ok {
		if r.processFuncLit(funcLit) {
			modified = true
		}
	}

	return modified
}

// processStmtList 递归处理语句列表
func (r *Rewrite) processStmtList(stmts []ast.Stmt) bool {
	modified := false
	for _, stmt := range stmts {
		if r.processStmt(stmt) {
			modified = true
		}
	}
	return modified
}

// processStmt 递归处理语句，查找go语句和函数字面量
func (r *Rewrite) processStmt(stmt ast.Stmt) bool {
	modified := false

	switch s := stmt.(type) {
	case *ast.GoStmt:
		if r.processGoStmt(s) {
			modified = true
		}
	case *ast.ExprStmt:
		// 处理表达式语句中的函数调用
		if callExpr, ok := s.X.(*ast.CallExpr); ok {
			modified = r.processCallExpr(callExpr)
		}
	case *ast.BlockStmt:
		modified = r.processStmtList(s.List)
	case *ast.IfStmt:
		if s.Init != nil {
			modified = r.processStmt(s.Init) || modified
		}
		if s.Body != nil {
			modified = r.processStmtList(s.Body.List) || modified
		}
		if s.Else != nil {
			modified = r.processStmt(s.Else) || modified
		}
	case *ast.ForStmt:
		if s.Init != nil {
			modified = r.processStmt(s.Init) || modified
		}
		if s.Body != nil {
			modified = r.processStmtList(s.Body.List) || modified
		}
	case *ast.RangeStmt:
		if s.Body != nil {
			modified = r.processStmtList(s.Body.List) || modified
		}
	case *ast.SwitchStmt:
		if s.Init != nil {
			modified = r.processStmt(s.Init) || modified
		}
		if s.Body != nil {
			for _, caseClause := range s.Body.List {
				if cc, ok := caseClause.(*ast.CaseClause); ok {
					modified = r.processStmtList(cc.Body) || modified
				}
			}
		}
	case *ast.TypeSwitchStmt:
		if s.Init != nil {
			modified = r.processStmt(s.Init) || modified
		}
		if s.Body != nil {
			for _, caseClause := range s.Body.List {
				if cc, ok := caseClause.(*ast.CaseClause); ok {
					modified = r.processStmtList(cc.Body) || modified
				}
			}
		}
	case *ast.SelectStmt:
		if s.Body != nil {
			for _, commClause := range s.Body.List {
				if cc, ok := commClause.(*ast.CommClause); ok {
					modified = r.processStmtList(cc.Body) || modified
				}
			}
		}
	}

	return modified
}

func (r *Rewrite) RewriteFile() {
	flag := false
	// 插入defer函数
	for _, item := range r.f.Decls {
		funcDel, ok := item.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if strings.ToLower(funcDel.Name.Name) == "string" {
			continue
		}

		// 检查函数是否有函数体，没有函数体的函数（如接口方法声明）跳过
		if funcDel.Body == nil {
			continue
		}

		// 为函数声明添加defer语句
		if r.addDeferToBody(funcDel.Body, funcDel.Type, funcDel.Recv) {
			flag = true
		}

		// 递归处理函数体中的所有语句，查找go语句
		if r.processStmtList(funcDel.Body.List) {
			flag = true
		}
	}
	if flag {
		// 插入import
		r.ImportFunctrace()
	}
	buf := &bytes.Buffer{}
	err := format.Node(buf, r.fset, r.f)
	if err != nil {
		fmt.Printf("rewrite found err:%s \n", err)
		return
	}
	if debug {
		fmt.Println(buf.String())
		return
	}
	if err = os.WriteFile(r.fullPath, buf.Bytes(), 0o666); err != nil {
		fmt.Printf("write %s error: %v\n", r.fullPath, err)
		return
	}
}
