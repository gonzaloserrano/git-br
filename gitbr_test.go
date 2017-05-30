package gitbr_test

import (
	"testing"

	gitbr "github.com/gonzaloserrano/git-br"
	"github.com/stretchr/testify/assert"
)

func TestOpenInvalidPathErrors(t *testing.T) {
	assert := assert.New(t)

	ui, err := gitbr.Open("foo")
	assert.Error(err)
	assert.Nil(ui)
}

func TestOpenCurrentRepository(t *testing.T) {
	assert := assert.New(t)

	ui, err := gitbr.Open("")
	assert.NoError(err)
	assert.NotNil(ui)
}
