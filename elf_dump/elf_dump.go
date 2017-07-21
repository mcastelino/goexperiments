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

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-f <template>] <filename> \n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "The template passed to the -%s option for FileHeader operates on a\n\n", "f")
		fmt.Fprintln(os.Stderr, tfortools.GenerateUsageUndecorated(elf.FileHeader{}))
		fmt.Fprintf(os.Stderr, "The template passed to the -%s option for ProgramHeaders operates on a\n\n", "f")
		fmt.Fprintln(os.Stderr, tfortools.GenerateUsageUndecorated([]elf.ProgHeader{}))
		fmt.Fprintf(os.Stderr, "The template passed to the -%s option for Sections operates on a\n\n", "f")
		fmt.Fprintln(os.Stderr, tfortools.GenerateUsageUndecorated([]elf.SectionHeader{}))
		fmt.Fprintln(os.Stderr, tfortools.TemplateFunctionHelp(cfg))
	}

	flag.StringVar(&code, "f", "", "string containing the template code to execute")

	cfg = tfortools.NewConfig(tfortools.OptAllFns)

	if err := cfg.AddCustomFn(getSections, "getSections", getSectionsHelp); err != nil {
		panic(err)
	}

	if err := cfg.AddCustomFn(getPrograms, "getPrograms", getProgramsHelp); err != nil {
		panic(err)
	}
}

const getSectionsHelp = "- getSections extracts all the sections from the elf \n"

func getSections(f *elf.File) []elf.SectionHeader {
	var sectionHeaders []elf.SectionHeader

	for _, s := range f.Sections {
		sectionHeaders = append(sectionHeaders, s.SectionHeader)
	}
	return sectionHeaders
}

const getProgramsHelp = "- getPrograms extracts all the sections from the elf \n"

func getPrograms(f *elf.File) []elf.ProgHeader {
	var progHeaders []elf.ProgHeader

	for _, p := range f.Progs {
		progHeaders = append(progHeaders, p.ProgHeader)
	}
	return progHeaders
}

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	filename := flag.Args()[0]
	if filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	r := bytes.NewReader(contents)

	f, err := elf.NewFile(r)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if code == "" {
		fmt.Printf("\nFile Header:\n")
		err = tfortools.OutputToTemplate(os.Stdout, "File", "{{println .}}", f.FileHeader, cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		fmt.Printf("\nProgram Headers:\n")
		tfortools.OutputToTemplate(os.Stdout, "Program headers", "{{ table (getPrograms .) }}", f, cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		fmt.Printf("\nSection Headers:\n")
		tfortools.OutputToTemplate(os.Stdout, "Sections", "{{ table (getSections .) }}", f, cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	err = tfortools.OutputToTemplate(os.Stdout, "f", code, f, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
