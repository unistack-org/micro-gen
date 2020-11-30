package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/unistack-org/micro/v3/logger"
)

func isGenerated(name string) bool {
	const (
		genCodeGenerated = "code generated"
		genDoNotEdit     = "do not edit"
		genAutoFile      = "autogenerated file"
	)

	markers := []string{genCodeGenerated, genDoNotEdit, genAutoFile}

	fileset := token.NewFileSet()
	syntax, err := parser.ParseFile(fileset, name, nil, parser.PackageClauseOnly|parser.ParseComments)
	if err != nil {
		return false
	}

	for _, comment := range syntax.Comments {
		for _, marker := range markers {
			if strings.Contains(strings.ToLower(comment.Text()), marker) {
				return true
			}
		}
	}

	return false
}

func cleanDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if dir == "vendor" {
			logger.Info("skip vendor dir")
			return filepath.SkipDir
		}
		if isGenerated(path) {
			logger.Infof("remove generated file: %s", path)
			err = os.Remove(path)
		}
		return err
	})
}
