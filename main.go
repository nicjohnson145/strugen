package main

import (
	lib "github.com/nicjohnson145/strugen/strugenlib"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
	"os"
	"path"
	"text/template"
)

func main() {
	if err := cmd().Execute(); err != nil {
		log.Fatal(err)
	}
}

func cmd() *cobra.Command {
	opts := strugenOpts{}

	root := &cobra.Command{
		Use:  "strugen",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			run(opts, args[0])
		},
	}
	root.Flags().StringSliceVarP(&opts.Types, "type", "t", []string{}, "Types to generate for")
	return root
}

type strugenOpts struct {
	Types []string
}

type renderVars struct {
	Pkg     string
	Structs []lib.Struct
}

func HasTagValue(tagValues string, query string) bool {
	return lo.Contains(strings.Split(tagValues, ","), query)
}

func run(opts strugenOpts, templateFile string) {
	// Parse the user supplied template
	funcs := template.FuncMap{
		"HasTagValue": HasTagValue,
	}
	tmpl, err := template.New(path.Base(templateFile)).Funcs(funcs).ParseFiles(templateFile)
	if err != nil {
		log.Fatalf("error parsing template: %v", err)
	}

	gen := lib.Generator{
		Types:   opts.Types,
		TagName: "strugen",
	}
	structs, pkgname, err := gen.FindStructs()
	if err != nil {
		log.Fatalf("error finding structs: %v", err)
	}

	f, err := os.Create(fileName())
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	err = tmpl.Execute(f, renderVars{
		Pkg:     pkgname,
		Structs: lo.Values(structs),
	})
	if err != nil {
		log.Fatal("error rendering template: %v", err)
	}

}

func fileName() string {
	return "zzz_generated_strugen.go"
}
