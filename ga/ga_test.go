package ga

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfidenceInterval95(t *testing.T) {
	var ci float64

	ci = confidenceInterval95(0, 0)
	assert.Equal(t, 0.0, ci)

	ci = confidenceInterval95(0, 1)
	assert.Equal(t, 0.0, ci)

	ci = confidenceInterval95(1, 0)
	assert.Equal(t, 0.0, ci)

	ci = confidenceInterval95(10, 100)
	assert.Equal(t, 116, int(ci))

	ci = confidenceInterval95(100, 1000)
	assert.Equal(t, 36, int(ci))
}
