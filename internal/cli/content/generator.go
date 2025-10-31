package content

// Generator handles Hugo markdown generation
type Generator struct {
	contentDir string
}

// NewGenerator creates a new content generator
func NewGenerator(contentDir string) *Generator {
	return &Generator{
		contentDir: contentDir,
	}
}
