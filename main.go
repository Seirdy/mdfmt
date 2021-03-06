package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	formatter "github.com/mdigger/goldmark-formatter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"gopkg.in/yaml.v3"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage of %s [flags] < input > output:\n", os.Args[0])
		flag.PrintDefaults()
	}
	var skipMetadata bool
	flag.BoolVar(&skipMetadata, "skipMetadata", false, "remove metadata front matter")
	flag.BoolVar(&formatter.SkipHTML, "skipHTML", false, "remove HTML blocks")
	flag.BoolVar(&formatter.LineBreak, "wrapLines", false, "hard wrap lines")
	flag.BoolVar(&formatter.STXHeader, "stxHeaders", false, "use STX headers")
	flag.Parse()

	// load markdown
	source, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	var out = os.Stdout

	// decode metadata if exists
	if bytes.HasPrefix(source, []byte("---\n")) {
		// search the end of metadata
		var (
			start = 4
			end   int
		)
	research:
		for _, marker := range []string{"\n---", "\n..."} {
			end = bytes.Index(source[start:], []byte(marker))
			if end != -1 {
				break
			}
		}
		// check find metadata
		if end != -1 {
			// parse yaml front matter
			var meta yaml.Node
			err = yaml.Unmarshal(source[4:start+end], &meta)
			if err != nil || len(meta.Content) != 1 {
				start += end + 4
				goto research
			}
			// skip metadata from source
			source = source[start+end+4:]
			if !skipMetadata {
				// rewrite metadata
				io.WriteString(out, "---\n")
				enc := yaml.NewEncoder(out)
				err = enc.Encode(meta.Content[0])
				enc.Close()
				if err != nil {
					log.Fatal(err)
				}
				io.WriteString(out, "---\n")
			}
		}
	}

	var md = goldmark.New(
		goldmark.WithRenderer(
			formatter.Markdown,
		),
		goldmark.WithParserOptions(
			parser.WithAttribute(),
		),
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Table,
			extension.Strikethrough,
			extension.DefinitionList,
			extension.Footnote,
			// attributes.Extension, // add block attributes support
			// lineblocks.Extension, // enable inline blocks
		),
	)

	// parse markdown and write reformatted source
	err = md.Convert(source, out)
	if err != nil {
		log.Fatal(err)
	}
}
