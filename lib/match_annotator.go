package rat

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/MathieuTurcotte/go-trie/gtrie"
)

type matchAnnotator struct {
	class    string
	trieRoot *gtrie.Node
	loading  chan bool
}

func NewMatchAnnotator(cmd, class string) Annotator {
	ma := &matchAnnotator{}

	ma.class = class
	ma.loading = make(chan bool)

	go ma.loadMatches(cmd)

	return ma
}

func (ma *matchAnnotator) loadMatches(cmd string) {
	defer close(ma.loading)

	command := exec.Command(os.Getenv("SHELL"), "-c", cmd)
	output, _ := command.Output()
	lines := strings.Split(string(output), "\n")
	matchStrings := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if len(trimmedLine) > 0 {
			matchStrings = append(matchStrings, trimmedLine)
		}
	}

	if len(matchStrings) > 0 {
		sort.Strings(matchStrings)
		ma.trieRoot, _ = gtrie.Create(matchStrings)
	}
}

func (ma *matchAnnotator) Annotate(reader BufferReader) <-chan Annotation {
	out := make(chan Annotation)

	go func() {
		defer close(out)

		var (
			pr     PositionedRune
			err    error
			start  BufferPoint
			buf    bytes.Buffer
			next   *gtrie.Node
			cursor *gtrie.Node
		)

		<-ma.loading

		if ma.trieRoot == nil {
			return
		}

		cursor = ma.trieRoot

		for {
			pr, err = reader.ReadPositionedRune()

			if err == io.EOF {
				break
			}

			if next = cursor.GetChild(pr.Rune()); next != nil {
				buf.WriteRune(pr.Rune())
				cursor = next

				if start == nil {
					start = pr.Pos()
				}

				if cursor.Terminal {
					out <- NewAnnotation(start, pr.Pos(), ma.class, buf.String())
				} else {
					continue
				}
			}

			buf.Reset()
			cursor = ma.trieRoot
			start = nil
		}
	}()

	return out
}
