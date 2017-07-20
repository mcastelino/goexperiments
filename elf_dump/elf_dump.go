package main

import (
	"bytes"
	"debug/elf"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/intel/tfortools"
)

var cfg *tfortools.Config
var code string //The format string

var fileHeader elf.FileHeader
var progHeaders []elf.ProgHeader
var sectionHeaders []elf.SectionHeader

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s \n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "The template passed to the -%s option for FileHeader operates on a\n\n", "f")
		fmt.Fprintln(os.Stderr, tfortools.GenerateUsageUndecorated(fileHeader))
		fmt.Fprintf(os.Stderr, "The template passed to the -%s option for ProgramHeaders operates on a\n\n", "f")
		fmt.Fprintln(os.Stderr, tfortools.GenerateUsageUndecorated(progHeaders))
		fmt.Fprintf(os.Stderr, "The template passed to the -%s option for Sections operates on a\n\n", "f")
		fmt.Fprintln(os.Stderr, tfortools.GenerateUsageUndecorated(sectionHeaders))
		fmt.Fprintln(os.Stderr, tfortools.TemplateFunctionHelp(cfg))
	}
	flag.StringVar(&code, "f", "", "string containing the template code to execute")
	cfg = tfortools.NewConfig(tfortools.OptAllFns)
}

func main() {
	filename := flag.String("filename", "", "file to dump")
	component := flag.String("component", "",
		"file component to dump (file|section|program), if unspecified dumps all possible headers")

	flag.Parse()

	contents, err := ioutil.ReadFile(*filename)
	if err != nil {
		log.Fatal(err)
	}

	r := bytes.NewReader(contents)

	f, err := elf.NewFile(r)
	if err != nil {
		log.Fatal(err)
	}

	fileHeader = f.FileHeader

	for _, p := range f.Progs {
		progHeaders = append(progHeaders, p.ProgHeader)
	}

	for _, s := range f.Sections {
		sectionHeaders = append(sectionHeaders, s.SectionHeader)
	}

	if *component == "" {
		if code == "" {
			fmt.Println("\nFile Header:")
			tfortools.OutputToTemplate(os.Stdout, "File", "{{println .}}", fileHeader, cfg)
			fmt.Printf("\nProgram Headers:")
			tfortools.OutputToTemplate(os.Stdout, "Program headers", "{{table .}}", progHeaders, cfg)
			fmt.Println("\nSection Headers:")
			tfortools.OutputToTemplate(os.Stdout, "Sections", "{{table .}}", sectionHeaders, cfg)
		} else {
			fmt.Println("\nFile Header:")
			tfortools.OutputToTemplate(os.Stdout, "File", code, fileHeader, cfg)
			fmt.Printf("\nProgram Headers:")
			tfortools.OutputToTemplate(os.Stdout, "Program headers", code, progHeaders, cfg)
			fmt.Println("\nSection Headers:")
			tfortools.OutputToTemplate(os.Stdout, "Sections", code, sectionHeaders, cfg)
		}
		return
	}

	switch *component {
	case "program":
		err = tfortools.OutputToTemplate(os.Stdout, "Program headers", code, progHeaders, cfg)
	case "file":
		err = tfortools.OutputToTemplate(os.Stdout, "File header", code, fileHeader, cfg)
	case "section":
		err = tfortools.OutputToTemplate(os.Stdout, "Sections", code, sectionHeaders, cfg)
	}

	if err != nil {
		fmt.Errorf("Unable to execute template : %v", err)
	}
}
