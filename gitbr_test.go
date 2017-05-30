package gitbr_test

import (
	"testing"

	gitbr "github.com/gonzaloserrano/git-br"
	"github.com/stretchr/testify/assert"
)

func TestOpenInvalidPathErrors(t *testing.T) {
	assert := assert.New(t)

	err := gitbr.Open("foo")
	assert.Error(err)
}
