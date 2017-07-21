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

		for _, f := range tmplFnTable {
			fmt.Fprintf(os.Stderr, "The template passed to the -%s option for %s operates on a\n\n", "f", f.name)
			fmt.Fprintln(os.Stderr, tfortools.GenerateUsageUndecorated(f.object))
		}
		fmt.Fprintln(os.Stderr, tfortools.TemplateFunctionHelp(cfg))
	}

	flag.StringVar(&code, "f", "", "string containing the template code to execute")

	cfg = tfortools.NewConfig(tfortools.OptAllFns)

	for _, f := range tmplFnTable {
		if err := cfg.AddCustomFn(f.fn, f.name, f.helpText); err != nil {
			panic(err)
		}
	}
}

type fnEntry struct {
	fn       interface{}
	object   interface{}
	name     string
	helpText string
}

// Table of custom elf parsing functions
// Always update this when you add more functions
var tmplFnTable []fnEntry = []fnEntry{
	{
		getFileHeader,
		[]elf.FileHeader{},
		"getFileHeader",
		getFileHeaderHelp,
	},
	{
		getSections,
		[]elf.Section{},
		"getSections",
		getSectionsHelp,
	},
	{
		getPrograms,
		[]elf.ProgHeader{},
		"getPrograms",
		getProgramsHelp,
	},
	{
		getSymbols,
		[]elf.Symbol{},
		"getSymbols",
		getSymbolsHelp,
	},
}

//Custom functions specific to elf parsing

//The file header is setup as slice so that the template function can be
//applied uniformly. There will always be a single file header
const getFileHeaderHelp = "- getFileHeader extracts the FileHeader from the elf \n"

func getFileHeader(f *elf.File) (fileHeaders []elf.FileHeader) {
	fileHeaders = append(fileHeaders, f.FileHeader)
	return
}

const getSectionsHelp = "- getSections extracts all the sections from the elf \n"

func getSections(f *elf.File) (sectionHeaders []elf.SectionHeader) {
	for _, s := range f.Sections {
		sectionHeaders = append(sectionHeaders, s.SectionHeader)
	}
	return
}

const getProgramsHelp = "- getPrograms extracts all the program headers from the elf \n"

func getPrograms(f *elf.File) (progHeaders []elf.ProgHeader) {
	for _, p := range f.Progs {
		progHeaders = append(progHeaders, p.ProgHeader)
	}
	return
}

const getSymbolsHelp = "- getSymbols extracts all the symbols from the elf \n"

func getSymbols(f *elf.File) (symbols []elf.Symbol) {
	symbols, err := f.Symbols()

	if err != nil {
		log.Printf("No symbols found [%v]", err)
	}
	return symbols
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

		fmt.Printf("\nSymbols:\n")
		tfortools.OutputToTemplate(os.Stdout, "Symbols", "{{ table (getSymbols .) }}", f, cfg)
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
