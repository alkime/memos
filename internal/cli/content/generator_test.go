package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGenerator(t *testing.T) {
	contentDir := "content/posts"

	generator := NewGenerator(contentDir)

	assert.NotNil(t, generator)
	assert.Equal(t, contentDir, generator.contentDir)
}
