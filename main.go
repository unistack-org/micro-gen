package main

import (
	"context"
	"flag"
	"fmt"
	"go/types"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/unistack-org/micro/v3/logger"
)

var (
	flagDir   string
	flagUrl   string
	flagForce bool
)

func init() {
	flag.StringVar(&flagDir, "dstdir", "", "place for generated files")
	flag.StringVar(&flagUrl, "url", "", "repo url path")
	flag.BoolVar(&flagForce, "force", false, "owerwrite files")
}

func main() {
	var err error
	flag.Parse()

	if len(flagDir) == 0 {
		if flagDir, err = os.Getwd(); err != nil {
			logger.Fatal(err)
		}
		logger.Infof("dstdir not specified, use current dir: %s", flagDir)
	}

	url := "https://github.com/unistack-org/micro"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cloneOpts := &git.CloneOptions{
		URL:          url,
		SingleBranch: true,
		Progress:     os.Stdout,
		Depth:        1,
	}
	if err := cloneOpts.Validate(); err != nil {
		logger.Fatal(err)
	}

	repo, err := git.CloneContext(ctx, memory.NewStorage(), nil, cloneOpts)
	if err != nil {
		logger.Fatal(err)
	}

	ref, err := repo.Head()
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Println(ref.Hash())
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		logger.Fatal(err)
	}

	tree, err := commit.Tree()
	if err != nil {
		logger.Fatal(err)
	}

	if err := os.MkdirAll(flagDir, os.FileMode(0755)); err != nil {
		logger.Fatal(err)
	}

	if err := cleanDir(flagDir); err != nil {
		logger.Fatal(err)
	}

	err = tree.Files().ForEach(func(file *object.File) error {
		if file == nil {
			return types.Error{Msg: "File pointer is empty"}
		}
		fmt.Printf("%#+v\n", file)
		return nil
	})

	if err != nil {
		logger.Fatal(err)
	}
}
