// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

// Observable assertion ratio checker.
//
// Constitution §1.2 requires that at least 60% of assertions in test
// files verify observable behaviour (actual outputs, side effects,
// state changes) rather than structural or metadata checks (nil/non-nil,
// type assertions, constructor-only existence checks).
//
// This tool analyses Go test files using AST and categorises each
// assertion call into:
//
//   OBSERVABLE  — the assertion checks a real value, error, or state.
//   STRUCTURAL  — the assertion only checks existence, type, or shape.
//
// Usage:
//   go run scripts/assertion_ratio_check.go ./...
//   go run scripts/assertion_ratio_check.go -min=0.60 ./...
//
// Exit code: 0 = ratio meets threshold, 1 = ratio below threshold.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	minRatio    = flag.Float64("min", 0.60, "minimum observable assertion ratio (0.0–1.0)")
	excludeDirs = flag.String("exclude", "vendor,node_modules,dist,build,.git,fixtures", "comma-separated directory names to skip")
	verbose     = flag.Bool("v", false, "verbose per-file output")
)

// observableAsserts are assertion functions that check actual values,
// errors, or state changes.
var observableAsserts = map[string]bool{
	"Equal":          true,
	"EqualValues":    true,
	"NotEqual":       true,
	"NotEqualValues": true,
	"True":           true, // checked: must be real condition, not 'true'
	"False":          true,
	"Error":          true,
	"NoError":        true,
	"ErrorIs":        true,
	"ErrorAs":        true,
	"Contains":       true,
	"NotContains":    true,
	"Greater":        true,
	"GreaterOrEqual": true,
	"Less":           true,
	"LessOrEqual":    true,
	"InDelta":        true,
	"InDeltaMapValues": true,
	"InEpsilon":      true,
	"JSONEq":         true,
	"YAMLEq":         true,
	"Regexp":         true,
	"NotRegexp":      true,
	"ElementsMatch":  true,
	"Subset":         true,
	"NotSubset":      true,
	"FileExists":     true,
	"DirExists":      true,
	"Eventually":     true,
	"Never":          true,
	"HTTPStatusCode": true,
	"HTTPBodyContains": true,
	"HTTPBodyNotContains": true,
	"HTTPError":      true,
	"HTTPSuccess":    true,
	"HTTPRedirect":   true,
}

// structuralAsserts are assertion functions that only check shape,
// existence, or type without verifying observable behaviour.
var structuralAsserts = map[string]bool{
	"Nil":         true,
	"NotNil":      true,
	"Empty":       true,
	"NotEmpty":    true,
	"IsType":      true,
	"Implements":  true,
	"IsIncreasing": true,
	"IsNonIncreasing": true,
	"IsDecreasing": true,
	"IsNonDecreasing": true,
	"Same":        true,
	"NotSame":     true,
	"Panics":      true,
	"NotPanics":   true,
	"Fail":        true,
	"FailNow":     true,
	"Fatal":       true,
	"Fatalf":      true,
	"Skip":        true,
	"SkipNow":     true,
	"Skipf":       true,
}

// uncertainAsserts are not counted toward either bucket because they
// can be used in either observable or structural contexts depending
// on arguments.
var uncertainAsserts = map[string]bool{
	"Len": true, // Len(t, obj, 3) could be either
}

type fileStats struct {
	path        string
	observable  int
	structural  int
	uncertain   int
	bluffCount  int // assert.True(t, true) etc.
}

func main() {
	flag.Parse()
	patterns := flag.Args()
	if len(patterns) == 0 {
		patterns = []string{"."}
	}

	excludeDirSet := make(map[string]struct{})
	for _, d := range strings.Split(*excludeDirs, ",") {
		excludeDirSet[strings.TrimSpace(d)] = struct{}{}
	}

	dirs, err := listPackageDirs(patterns, excludeDirSet)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing packages: %v\n", err)
		os.Exit(1)
	}

	var files []string
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, "_test.go") {
				continue
			}
			files = append(files, filepath.Join(dir, name))
		}
	}

	if *verbose {
		fmt.Printf("Scanning %d test files...\n", len(files))
	}

	var allStats []fileStats
	var totalObservable, totalStructural, totalUncertain, totalBluff int

	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, content, parser.ParseComments)
		if err != nil {
			continue
		}

		st := analyseFile(f)
		st.path = path
		allStats = append(allStats, st)
		totalObservable += st.observable
		totalStructural += st.structural
		totalUncertain += st.uncertain
		totalBluff += st.bluffCount
	}

	totalCounted := totalObservable + totalStructural
	var ratio float64
	if totalCounted > 0 {
		ratio = float64(totalObservable) / float64(totalCounted)
	}

	if *verbose {
		for _, st := range allStats {
			fileTotal := st.observable + st.structural
			var fileRatio float64
			if fileTotal > 0 {
				fileRatio = float64(st.observable) / float64(fileTotal)
			}
			fmt.Printf("  %s: %d observable / %d structural / %d uncertain / %.0f%%\n",
				st.path, st.observable, st.structural, st.uncertain, fileRatio*100)
		}
		fmt.Println()
	}

	fmt.Printf("Observable assertion ratio: %d observable / %d structural = %.1f%% (threshold: %.0f%%)\n",
		totalObservable, totalStructural, ratio*100, *minRatio*100)

	if totalBluff > 0 {
		fmt.Printf("VACUOUS ASSERTIONS FOUND: %d (assert.True(t, true) etc.)\n", totalBluff)
	}

	if totalCounted == 0 {
		fmt.Println("No assertions found.")
		os.Exit(1)
	}

	if ratio < *minRatio {
		fmt.Fprintf(os.Stderr, "ERROR: observable assertion ratio %.1f%% is below threshold %.0f%%\n",
			ratio*100, *minRatio*100)
		os.Exit(1)
	}

	if totalBluff > 0 {
		fmt.Fprintf(os.Stderr, "ERROR: %d vacuous assertion(s) found\n", totalBluff)
		os.Exit(1)
	}

	fmt.Println("PASS: observable assertion ratio meets threshold and no vacuous assertions found.")
	os.Exit(0)
}

// analyseFile walks a single test file AST and categorises assertions.
func analyseFile(f *ast.File) fileStats {
	var st fileStats
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		funcName := extractAssertName(call)
		if funcName == "" {
			return true
		}

		// Detect vacuous assertions like assert.True(t, true).
		if isVacuousAssertion(funcName, call) {
			st.bluffCount++
			return true
		}

		if observableAsserts[funcName] {
			st.observable++
		} else if structuralAsserts[funcName] {
			st.structural++
		} else if uncertainAsserts[funcName] {
			st.uncertain++
		}
		// Unknown assertion names are ignored (custom helpers, etc.)
		return true
	})
	return st
}

// extractAssertName returns the assertion function name if the call is
// to a known test assertion package (testify assert/require, or
// standard testing.T methods).
func extractAssertName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		// e.g., assert.Equal, require.NoError
		if ident, ok := fn.X.(*ast.Ident); ok {
			pkg := ident.Name
			if pkg == "assert" || pkg == "require" || pkg == "suite" {
				return fn.Sel.Name
			}
		}
	case *ast.Ident:
		// Direct call if package was dot-imported (rare, but handle it)
		return fn.Name
	}
	return ""
}

// isVacuousAssertion detects the canonical bluff patterns:
//   assert.True(t, true)
//   assert.False(t, false)
//   assert.NotNil(t, nil)
//   assert.Nil(t, nil)  (always passes, vacuous)
func isVacuousAssertion(name string, call *ast.CallExpr) bool {
	args := call.Args
	if len(args) < 2 {
		return false
	}
	// First arg after *testing.T is the value being asserted.
	valArg := args[1]

	switch name {
	case "True":
		return isBoolLiteral(valArg, true)
	case "False":
		return isBoolLiteral(valArg, false)
	case "NotNil":
		return isNilLiteral(valArg)
	case "Nil":
		return isNilLiteral(valArg)
	}
	return false
}

func isBoolLiteral(expr ast.Expr, val bool) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	if val {
		return ident.Name == "true"
	}
	return ident.Name == "false"
}

func isNilLiteral(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "nil"
}

// listPackageDirs expands Go package patterns into a deduplicated list of
// package directory paths. Excludes directories matching excludeDirSet.
func listPackageDirs(patterns []string, excludeDirSet map[string]struct{}) ([]string, error) {
	args := append([]string{"list", "-f", "{{.Dir}}"}, patterns...)
	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var dirs []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		dir := strings.TrimSpace(scanner.Text())
		if dir == "" {
			continue
		}
		skip := false
		for part := range excludeDirSet {
			if strings.Contains(dir, string(filepath.Separator)+part+string(filepath.Separator)) ||
				strings.HasSuffix(dir, string(filepath.Separator)+part) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if _, ok := seen[dir]; !ok {
			seen[dir] = struct{}{}
			dirs = append(dirs, dir)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		if len(dirs) == 0 {
			return nil, err
		}
	}

	return dirs, nil
}
