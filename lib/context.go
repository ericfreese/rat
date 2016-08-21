package rat

import (
	"fmt"
	"strings"
)

func InterpolateContext(str string, ctx Context) string {
	for k, v := range ctx {
		str = strings.Replace(str, fmt.Sprintf("%%(%s)", k), v, -1)
	}

	return str
}

func MergeContext(orig, extra Context) Context {
	merged := Context{}

	for k, v := range orig {
		merged[k] = v
	}

	for k, v := range extra {
		merged[k] = v
	}

	return merged
}
