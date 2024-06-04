package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetBranchOptions(t *testing.T) {
	assert := assert.New(t)
	cases := []string{"feat", "bug", "docs", "refactor", "perf", "ci", ""}
	for _, c := range cases {
		t.Run(fmt.Sprintf("Case %s", c), func(t *testing.T) {
			opts := getBranchOptions(c)
			assert.Len(opts, 7)
			assert.Equal(c, opts[0].Value)
		})
	}
}
