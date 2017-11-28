package rat

import (
	"fmt"
	"os"
)

type Context map[string]string

func ContextEnvironment(ctx Context) []string {
	env := os.Environ()

	for k, v := range ctx {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
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
