package rewrite

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRewrite(t *testing.T) {
	// 创建临时测试文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func testFunc() {
	fmt.Println("test")
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rewrite, err := NewRewrite(testFile)
	if err != nil {
		t.Fatalf("NewRewrite failed: %v", err)
	}

	if rewrite == nil {
		t.Fatal("NewRewrite returned nil")
	}

	if rewrite.fullPath != testFile {
		t.Errorf("Expected fullPath %s, got %s", testFile, rewrite.fullPath)
	}

	if rewrite.fset == nil {
		t.Error("fset should not be nil")
	}

	if rewrite.f == nil {
		t.Error("f should not be nil")
	}
}

func TestGenTraceParams(t *testing.T) {
	rewrite := &Rewrite{}

	tests := []struct {
		name     string
		funcType *ast.FuncType
		recv     *ast.FieldList
		expected int // 期望的参数数量
	}{
		{
			name: "no parameters",
			funcType: &ast.FuncType{
				Params: &ast.FieldList{},
			},
			expected: 0,
		},
		{
			name: "with parameters",
			funcType: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{{Name: "a"}},
						},
						{
							Names: []*ast.Ident{{Name: "b"}},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "with receiver",
			funcType: &ast.FuncType{
				Params: &ast.FieldList{},
			},
			recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: "r"}},
					},
				},
			},
			expected: 1,
		},
		{
			name: "with underscore parameters",
			funcType: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{{Name: "_"}},
						},
						{
							Names: []*ast.Ident{{Name: "valid"}},
						},
					},
				},
			},
			expected: 1, // 只有valid参数会被包含
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewrite.genTraceParams(tt.funcType, tt.recv)
			if len(result) != tt.expected {
				t.Errorf("Expected %d parameters, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestGenDefer(t *testing.T) {
	rewrite := &Rewrite{}

	elts := []ast.Expr{
		&ast.BasicLit{Kind: token.STRING, Value: `"param1"`},
		&ast.BasicLit{Kind: token.STRING, Value: `"param2"`},
	}

	deferStmt := rewrite.genDefer(elts)
	if deferStmt == nil {
		t.Fatal("genDefer returned nil")
	}

	// 检查是否为defer语句
	if deferStmt.Call == nil {
		t.Error("defer statement should have a call")
	}

	// 检查函数调用结构
	callExpr, ok := deferStmt.Call.Fun.(*ast.CallExpr)
	if !ok {
		t.Error("defer call should be a CallExpr")
	}

	// 检查是否为functrace.Trace调用
	selector, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		t.Error("defer call should have a SelectorExpr")
	}

	if ident, ok := selector.X.(*ast.Ident); !ok || ident.Name != "functrace" {
		t.Error("defer call should call functrace.Trace")
	}

	if selector.Sel.Name != "Trace" {
		t.Error("defer call should call functrace.Trace")
	}
}

func TestHasSameDefer(t *testing.T) {
	rewrite := &Rewrite{}

	tests := []struct {
		name     string
		funcDecl *ast.FuncDecl
		expected bool
	}{
		{
			name: "no body",
			funcDecl: &ast.FuncDecl{
				Body: nil,
			},
			expected: false,
		},
		{
			name: "no defer statements",
			funcDecl: &ast.FuncDecl{
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ExprStmt{},
					},
				},
			},
			expected: false,
		},
		{
			name: "has functrace defer",
			funcDecl: &ast.FuncDecl{
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.DeferStmt{
							Call: &ast.CallExpr{
								Fun: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "functrace"},
										Sel: &ast.Ident{Name: "Trace"},
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewrite.HasSameDefer(tt.funcDecl)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestHasSameDeferInBody(t *testing.T) {
	rewrite := &Rewrite{}

	tests := []struct {
		name     string
		body     *ast.BlockStmt
		expected bool
	}{
		{
			name:     "nil body",
			body:     nil,
			expected: false,
		},
		{
			name: "no functrace defer",
			body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.DeferStmt{
						Call: &ast.CallExpr{
							Fun: &ast.Ident{Name: "otherFunc"},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "has functrace defer",
			body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.DeferStmt{
						Call: &ast.CallExpr{
							Fun: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "functrace"},
									Sel: &ast.Ident{Name: "Trace"},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewrite.hasSameDeferInBody(tt.body)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAddDeferToBody(t *testing.T) {
	rewrite := &Rewrite{}

	tests := []struct {
		name     string
		body     *ast.BlockStmt
		funcType *ast.FuncType
		recv     *ast.FieldList
		expected bool
	}{
		{
			name:     "nil body",
			body:     nil,
			funcType: &ast.FuncType{},
			expected: false,
		},
		{
			name: "empty body",
			body: &ast.BlockStmt{
				List: []ast.Stmt{},
			},
			funcType: &ast.FuncType{},
			expected: true,
		},
		{
			name: "body with existing functrace defer",
			body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.DeferStmt{
						Call: &ast.CallExpr{
							Fun: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "functrace"},
									Sel: &ast.Ident{Name: "Trace"},
								},
							},
						},
					},
				},
			},
			funcType: &ast.FuncType{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialLen := 0
			if tt.body != nil {
				initialLen = len(tt.body.List)
			}

			result := rewrite.addDeferToBody(tt.body, tt.funcType, tt.recv)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}

			// 如果应该添加defer语句，检查是否真的添加了
			if tt.expected && tt.body != nil {
				if len(tt.body.List) != initialLen+1 {
					t.Errorf("Expected body to have %d statements, got %d", initialLen+1, len(tt.body.List))
				}

				// 检查第一个语句是否为defer语句
				if _, ok := tt.body.List[0].(*ast.DeferStmt); !ok {
					t.Error("First statement should be a defer statement")
				}
			}
		})
	}
}

func TestProcessFuncLit(t *testing.T) {
	rewrite := &Rewrite{}

	tests := []struct {
		name     string
		funcLit  *ast.FuncLit
		expected bool
	}{
		{
			name: "nil body",
			funcLit: &ast.FuncLit{
				Body: nil,
			},
			expected: false,
		},
		{
			name: "empty body",
			funcLit: &ast.FuncLit{
				Type: &ast.FuncType{},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{},
				},
			},
			expected: true,
		},
		{
			name: "body with existing functrace defer",
			funcLit: &ast.FuncLit{
				Type: &ast.FuncType{},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.DeferStmt{
							Call: &ast.CallExpr{
								Fun: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "functrace"},
										Sel: &ast.Ident{Name: "Trace"},
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewrite.processFuncLit(tt.funcLit)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestProcessGoStmt(t *testing.T) {
	rewrite := &Rewrite{}

	tests := []struct {
		name     string
		goStmt   *ast.GoStmt
		expected bool
	}{
		{
			name: "direct func literal",
			goStmt: &ast.GoStmt{
				Call: &ast.CallExpr{
					Fun: &ast.FuncLit{
						Type: &ast.FuncType{},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "function call with func literal parameter",
			goStmt: &ast.GoStmt{
				Call: &ast.CallExpr{
					Fun: &ast.Ident{Name: "someFunc"},
					Args: []ast.Expr{
						&ast.FuncLit{
							Type: &ast.FuncType{},
							Body: &ast.BlockStmt{
								List: []ast.Stmt{},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "function call without func literal",
			goStmt: &ast.GoStmt{
				Call: &ast.CallExpr{
					Fun: &ast.Ident{Name: "someFunc"},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: `"test"`},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewrite.processGoStmt(tt.goStmt)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestProcessCallExpr(t *testing.T) {
	rewrite := &Rewrite{}

	tests := []struct {
		name     string
		callExpr *ast.CallExpr
		expected bool
	}{
		{
			name: "no func literal parameters",
			callExpr: &ast.CallExpr{
				Fun: &ast.Ident{Name: "someFunc"},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: `"test"`},
				},
			},
			expected: false,
		},
		{
			name: "with func literal parameter",
			callExpr: &ast.CallExpr{
				Fun: &ast.Ident{Name: "someFunc"},
				Args: []ast.Expr{
					&ast.FuncLit{
						Type: &ast.FuncType{},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "nested function call",
			callExpr: &ast.CallExpr{
				Fun: &ast.Ident{Name: "outerFunc"},
				Args: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.Ident{Name: "innerFunc"},
						Args: []ast.Expr{
							&ast.FuncLit{
								Type: &ast.FuncType{},
								Body: &ast.BlockStmt{
									List: []ast.Stmt{},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewrite.processCallExpr(tt.callExpr)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRewriteFile_Integration(t *testing.T) {
	// 创建临时测试文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

import "fmt"

func main() {
	fmt.Println("Hello")
}

func testFunc(a, b int) string {
	return "test"
}

func (r *Receiver) method() {
	go func() {
		fmt.Println("goroutine")
	}()
	
	go someFunction(func() {
		fmt.Println("nested")
	})
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rewrite, err := NewRewrite(testFile)
	if err != nil {
		t.Fatalf("NewRewrite failed: %v", err)
	}

	// 执行重写
	rewrite.RewriteFile()

	// 读取重写后的文件
	rewrittenContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read rewritten file: %v", err)
	}

	contentStr := string(rewrittenContent)

	// 检查是否添加了import
	if !strings.Contains(contentStr, "github.com/toheart/functrace") {
		t.Error("Expected functrace import to be added")
	}

	// 检查是否在函数中添加了defer语句
	if !strings.Contains(contentStr, "defer functrace.Trace") {
		t.Error("Expected defer functrace.Trace to be added")
	}

	// 检查是否在goroutine中添加了defer语句
	if !strings.Contains(contentStr, "go func()") {
		t.Error("Expected goroutine to be preserved")
	}
}

func TestRewriteFile_StringMethod(t *testing.T) {
	// 创建临时测试文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func (r *Receiver) String() string {
	return "receiver"
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rewrite, err := NewRewrite(testFile)
	if err != nil {
		t.Fatalf("NewRewrite failed: %v", err)
	}

	// 执行重写
	rewrite.RewriteFile()

	// 读取重写后的文件
	rewrittenContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read rewritten file: %v", err)
	}

	contentStr := string(rewrittenContent)

	// String方法不应该被修改
	if strings.Contains(contentStr, "defer functrace.Trace") {
		t.Error("String method should not be modified")
	}
}

func TestRewriteFile_InterfaceMethod(t *testing.T) {
	// 创建临时测试文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

type Interface interface {
	Method()
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rewrite, err := NewRewrite(testFile)
	if err != nil {
		t.Fatalf("NewRewrite failed: %v", err)
	}

	// 执行重写
	rewrite.RewriteFile()

	// 读取重写后的文件
	rewrittenContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read rewritten file: %v", err)
	}

	contentStr := string(rewrittenContent)

	// 接口方法不应该被修改
	if strings.Contains(contentStr, "defer functrace.Trace") {
		t.Error("Interface method should not be modified")
	}
}

func TestRewriteDir(t *testing.T) {
	// 创建临时目录结构
	tmpDir := t.TempDir()

	// 创建子目录
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// 创建vendor目录
	vendorDir := filepath.Join(tmpDir, "vendor")
	err = os.Mkdir(vendorDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create vendor dir: %v", err)
	}

	// 创建测试文件
	files := map[string]string{
		filepath.Join(tmpDir, "main.go"): `package main

func main() {
	fmt.Println("main")
}`,
		filepath.Join(subDir, "helper.go"): `package helper

func helper() {
	fmt.Println("helper")
}`,
		filepath.Join(vendorDir, "vendor.go"): `package vendor

func vendor() {
	fmt.Println("vendor")
}`,
		filepath.Join(tmpDir, "test_test.go"): `package main

func TestMain(t *testing.T) {
	t.Log("test")
}`,
		filepath.Join(tmpDir, "non_go.txt"): `This is not a Go file`,
	}

	for path, content := range files {
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	// 执行目录重写
	RewriteDir(tmpDir)

	// 检查文件是否被修改
	mainContent, err := os.ReadFile(filepath.Join(tmpDir, "main.go"))
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	if !strings.Contains(string(mainContent), "defer functrace.Trace") {
		t.Error("main.go should be modified")
	}

	helperContent, err := os.ReadFile(filepath.Join(subDir, "helper.go"))
	if err != nil {
		t.Fatalf("Failed to read helper.go: %v", err)
	}

	if !strings.Contains(string(helperContent), "defer functrace.Trace") {
		t.Error("helper.go should be modified")
	}

	// vendor目录应该被跳过
	vendorContent, err := os.ReadFile(filepath.Join(vendorDir, "vendor.go"))
	if err != nil {
		t.Fatalf("Failed to read vendor.go: %v", err)
	}

	if strings.Contains(string(vendorContent), "defer functrace.Trace") {
		t.Error("vendor.go should not be modified")
	}

	// 测试文件应该被跳过
	testContent, err := os.ReadFile(filepath.Join(tmpDir, "test_test.go"))
	if err != nil {
		t.Fatalf("Failed to read test_test.go: %v", err)
	}

	if strings.Contains(string(testContent), "defer functrace.Trace") {
		t.Error("test_test.go should not be modified")
	}
}
