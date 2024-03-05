// microb is a frontend for buildkit to build docker images out of python projects.
//
// It reads the pyproject.toml file and generates a Dockerfile and then builds
// the image using buildkit.
// This file is inspired by mopy: https://github.com/cmdjulian/mopy/blob/main/cmd/mopy/main.go
package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"

	"github.com/charbonats/microbuild/v1/config"
	"github.com/charbonats/microbuild/v1/dockerfile"
	microbllb "github.com/charbonats/microbuild/v1/llb"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	"github.com/moby/buildkit/frontend/gateway/grpcclient"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/pkg/errors"
)

var filename string
var app string
var outputLLB bool
var outputDockerfile bool
var buildkit bool

func main() {
	flag.BoolVar(&outputLLB, "llb", false, "print llb to stdout")
	flag.BoolVar(&outputDockerfile, "dockerfile", false, "print equivalent Dockerfile to stdout")
	flag.BoolVar(&buildkit, "buildkit", true, "establish connection to buildkit and issue build")
	flag.StringVar(&filename, "filename", "pyproject.toml", "the pyproject.toml to build from")
	flag.StringVar(&app, "app", "", "the app to build")
	flag.Parse()

	// Display the dockerfile if requested
	if outputDockerfile {
		if err := printDockerfile(filename, app, os.Stdout); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	// Display the LLB if requested
	if outputLLB {
		if err := printLlb(filename, app, os.Stdout); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Build the image if requested
	if buildkit {
		if err := grpcclient.RunFromEnvironment(appcontext.Context(), microbllb.Build); err != nil {
			log.Fatal(err)
		}
	}
}

// printDockerfile prints the Dockerfile to the given writer
func printDockerfile(filename string, app string, out io.Writer) error {
	c, err := config.NewConfigFromFile(filename, &config.Options{Target: app})
	if err != nil {
		return errors.Wrap(err, "opening pyproject.toml")
	}
	dockerfile := dockerfile.Microb2Dockerfile(c, nil)
	out.Write([]byte(dockerfile))
	return nil
}

// printLlb prints the LLB to the given writer
func printLlb(filename string, app string, out io.Writer) error {
	c, err := config.NewConfigFromFile(filename, &config.Options{Target: app})
	if err != nil {
		return errors.Wrap(err, "opening pyproject.toml")
	}
	dockerfile := dockerfile.Microb2Dockerfile(c, nil)
	st, _, _, _ := dockerfile2llb.Dockerfile2LLB(context.TODO(), []byte(dockerfile), dockerfile2llb.ConvertOpt{})
	dt, err := st.Marshal(context.Background())
	if err != nil {
		return errors.Wrap(err, "marshaling llb state")
	}

	return llb.WriteTo(dt, out)
}
