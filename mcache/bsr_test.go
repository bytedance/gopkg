package mcache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBsr(t *testing.T) {
	assert.Equal(t, bsr(4), 2)
	assert.Equal(t, bsr(24), 4)
	assert.Equal(t, bsr((1<<10)-1), 9)
	assert.Equal(t, bsr((1<<30)+(1<<19)+(1<<16)+(1<<1)), 30)
}
