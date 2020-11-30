package main

import (
	"context"
	"flag"
	"fmt"
	"go/types"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/unistack-org/micro/v3/logger"
)

var (
	flagDir string
	flagUrl string
)

func init() {
	flag.StringVar(&flagDir, "dstdir", "", "place for generated files")
	flag.StringVar(&flagUrl, "url", "", "repo url path")
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	u, err := url.Parse(flagUrl)
	if err != nil {
		logger.Fatal(err)
	}

	var rev string
	if idx := strings.Index(u.Path, "@"); idx > 0 {
		rev = u.Path[idx+1:]
	}

	cloneOpts := &git.CloneOptions{
		URL:      flagUrl,
		Progress: os.Stdout,
	}

	if len(rev) == 0 {
		cloneOpts.SingleBranch = true
		cloneOpts.Depth = 1
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
			return types.Error{Msg: "file pointer is empty"}
		}

		fmode, err := file.Mode.ToOSFileMode()
		if err != nil {
			return err
		}

		switch file.Mode {
		case filemode.Executable:
			return writeFile(file, flagDir, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fmode)
		case filemode.Dir:
			return os.MkdirAll(filepath.Join(flagDir, file.Name), fmode)
		case filemode.Regular:
			return writeFile(file, flagDir, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fmode)
		default:
			return fmt.Errorf("unsupported filetype %v for %s", file.Mode, file.Name)
		}

		return nil
	})

	if err != nil {
		logger.Fatal(err)
	}
}
