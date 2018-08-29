package bridge

import (
	"testing"
	dockerapi "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
)

func buildContainerDetails() *dockerapi.Container {
	containerConfig := dockerapi.Config{
		Hostname: "testbox-1",
		User:     "testuser",
	}

	return &dockerapi.Container{
		ID:     "test-id",
		Config: &containerConfig,
		Image:  "configs/test",
	}
}

func TestEvaluateTemplateTags(t *testing.T) {
	s := "a-{{.Config.Status}}"
	result := EvaluateTemplateTags(&s, nil)

	assert.Equal(t, result, "a-{{.Config.Status}}")
}

func TestEvaluateTemplateTags_invalidTemplate(t *testing.T) {
	s := "a-{{.Config.Status}"
	result := EvaluateTemplateTags(&s, buildContainerDetails())

	assert.Equal(t, result, "a-{{.Config.Status}")
}

func TestEvaluateTemplateTags_validTemplate(t *testing.T) {
	s := "a-{{.Config.Hostname}}"
	result := EvaluateTemplateTags(&s, buildContainerDetails())

	assert.Equal(t, result, "a-testbox-1")
}

func TestEvaluateTemplateTags_validTemplateMultipletags(t *testing.T) {
	s := "a-{{.Config.Hostname}},b-{{.Config.User}},c-{{.Image}}"
	result := EvaluateTemplateTags(&s, buildContainerDetails())

	assert.Equal(t, result, "a-testbox-1,b-testuser,c-configs/test")
}

func TestCombineTags(t *testing.T) {
	res := combineTags()
	assert.Equal(t, 0, len(res))
}

func TestCombineTags_version2(t *testing.T) {
	res := combineTags("a-testbox-1,b-testuser,c-configs/test")
	assert.Equal(t, []string{"c-configs/test", "b-testuser", "a-testbox-1"}, res)
}
