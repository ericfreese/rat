package rat

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/MathieuTurcotte/go-trie/gtrie"
)

type Annotation interface {
	Start() int
	End() int
	Class() string
	Val() string
}

type Annotations interface {
	Add(Annotation)
	Intersecting(Region) []Annotation
	Len() int
}

type Annotator interface {
	Annotate(io.Reader) <-chan Annotation
}

type annotation struct {
	start int
	end   int
	class string
	val   string
}

func NewAnnotation(start, end int, class string, val string) Annotation {
	return &annotation{start, end, class, val}
}

func (a *annotation) Start() int {
	return a.start
}

func (a *annotation) End() int {
	return a.end
}

func (a *annotation) Class() string {
	return a.class
}

func (a *annotation) Val() string {
	return a.val
}

type annotations struct {
	annotations []Annotation
}

func NewAnnotations() Annotations {
	a := &annotations{}

	a.annotations = make([]Annotation, 0, 8)

	return a
}

func (a *annotations) Add(annotation Annotation) {
	a.annotations = append(a.annotations, annotation)
}

func (a *annotations) Intersecting(r Region) []Annotation {
	annotations := make([]Annotation, 0, 0)

	for _, annotation := range a.annotations {
		if annotation.Start() < r.End() && annotation.End() > r.Start() {
			annotations = append(annotations, annotation)
		}
	}

	return annotations
}

func (a *annotations) Len() int {
	return len(a.annotations)
}

type matchAnnotator struct {
	class    string
	trieRoot *gtrie.Node
	loading  chan bool
}

func NewMatchAnnotator(cmd, class string, ctx Context) Annotator {
	ma := &matchAnnotator{}

	ma.class = class
	ma.loading = make(chan bool)

	go ma.loadMatches(cmd, ctx)

	return ma
}

func (ma *matchAnnotator) loadMatches(cmd string, ctx Context) {
	defer close(ma.loading)

	command := exec.Command(os.Getenv("SHELL"), "-c", cmd)
	command.Env = ContextEnvironment(ctx)
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

func (ma *matchAnnotator) Annotate(rd io.Reader) <-chan Annotation {
	out := make(chan Annotation)

	go func() {
		defer close(out)

		var (
			r         rune
			size      int
			err       error
			offset    int
			start     int
			buf       bytes.Buffer
			next      *gtrie.Node
			cursor    *gtrie.Node
			candidate Annotation
		)

		<-ma.loading

		if ma.trieRoot == nil {
			return
		}

		cursor = ma.trieRoot

		bufrd := bufio.NewReader(rd)

		for {
			r, size, err = bufrd.ReadRune()

			if size == 0 && err == io.EOF {
				break
			}

			offset = offset + size

			if next = cursor.GetChild(r); next != nil {
				buf.WriteRune(r)
				cursor = next

				if start == -1 {
					start = offset - size
				}

				if cursor.Terminal {
					if cursor.HasChildren() {
						candidate = NewAnnotation(start, offset+size, ma.class, buf.String())
						continue
					} else {
						out <- NewAnnotation(start, offset+size, ma.class, buf.String())
					}
				} else {
					continue
				}
			} else if candidate != nil {
				out <- candidate
			}

			candidate = nil
			buf.Reset()
			cursor = ma.trieRoot
			start = -1
		}
	}()

	return out
}

type regexAnnotator struct {
	class string
	regex *regexp.Regexp
}

func NewRegexAnnotator(regex, class string) Annotator {
	ra := &regexAnnotator{}

	ra.class = class
	ra.regex = regexp.MustCompile(regex)

	return ra
}

func (ra *regexAnnotator) Annotate(rd io.Reader) <-chan Annotation {
	out := make(chan Annotation)

	go func() {
		defer close(out)

		if bytes, err := ioutil.ReadAll(rd); err == nil {
			for _, match := range ra.regex.FindAllIndex(bytes, -1) {
				out <- NewAnnotation(match[0], match[1], ra.class, string(bytes[match[0]:match[1]]))
			}
		}
	}()

	return out
}

type externalAnnotator struct {
	class string
	cmd   string
	ctx   Context
}

func NewExternalAnnotator(cmd, class string, ctx Context) Annotator {
	ea := &externalAnnotator{}

	ea.class = class
	ea.cmd = cmd
	ea.ctx = ctx

	return ea
}

func (ea *externalAnnotator) readUint64(rd io.Reader) (uint64, error) {
	buf := make([]byte, 8)

	if _, err := io.ReadFull(rd, buf); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(buf), nil
}

func (ea *externalAnnotator) Annotate(rd io.Reader) <-chan Annotation {
	out := make(chan Annotation)

	env := ContextEnvironment(ea.ctx)
	env = append(env, fmt.Sprintf("PATH=%s:%s", annotatorsDir, os.Getenv("PATH")))

	cmd := exec.Command(os.Getenv("SHELL"), "-c", ea.cmd)
	cmd.Env = env
	cmd.Stdin = rd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		close(out)
		return out
	}

	if err := cmd.Start(); err != nil {
		close(out)
		return out
	}

	go func() {
		defer close(out)

		for {
			start, err := ea.readUint64(stdout)
			if err != nil {
				return
			}

			end, err := ea.readUint64(stdout)
			if err != nil {
				return
			}

			lenValue, _ := ea.readUint64(stdout)
			if err != nil {
				return
			}

			value := make([]byte, lenValue)
			_, err = io.ReadFull(stdout, value)
			if err != nil {
				return
			}

			out <- NewAnnotation(
				int(start),
				int(end),
				ea.class,
				string(value),
			)
		}
	}()

	return out
}
