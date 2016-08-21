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

type matchParser struct {
	class    string
	trieRoot *gtrie.Node
	loading  chan bool
}

func NewMatchParser(cmd, class string) Annotator {
	mp := &matchParser{}

	mp.class = class
	mp.loading = make(chan bool)

	go mp.loadMatches(cmd)

	return mp
}

func (mp *matchParser) loadMatches(cmd string) {
	defer close(mp.loading)

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
		mp.trieRoot, _ = gtrie.Create(matchStrings)
	}
}

func (mp *matchParser) Annotate(reader BufferReader) <-chan Annotation {
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

		<-mp.loading

		if mp.trieRoot == nil {
			return
		}

		cursor = mp.trieRoot

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
					out <- NewAnnotation(start, pr.Pos(), mp.class, buf.String())
				} else {
					continue
				}
			}

			buf.Reset()
			cursor = mp.trieRoot
			start = nil
		}
	}()

	return out
}
