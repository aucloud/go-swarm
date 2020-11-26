package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullVersion(t *testing.T) {
	assert := assert.New(t)

	expected := fmt.Sprintf("%s-%s@%s", Version, Build, Commit)
	assert.Equal(expected, FullVersion())
}
