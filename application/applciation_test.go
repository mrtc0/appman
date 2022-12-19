package application

import (
	"log"
	"strings"
	"testing"

	"github.com/mitchellh/go-ps"
	"github.com/mrtc0/appman/application/config"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	conf := config.ApplicationConfig{
		Name:         "sleep",
		StartCommand: []string{"sleep", "10"},
	}

	logger := NewLogger(conf.Name, log.Writer())
	app := NewApplication(conf, logger)

	app.Start()

	processes, _ := ps.Processes()
	exists := false
	for _, p := range processes {
		if strings.Contains(p.Executable(), "sleep") {
			exists = true
		}
	}

	assert.True(t, exists)
	assert.Equal(t, app.status, Running)
	app.Stop()
}

func TestStop(t *testing.T) {
	conf := config.ApplicationConfig{
		Name:         "sleep",
		StartCommand: []string{"sleep", "10"},
	}

	logger := NewLogger(conf.Name, log.Writer())
	app := NewApplication(conf, logger)

	app.Start()
	app.Stop()

	processes, _ := ps.Processes()
	exists := false
	for _, p := range processes {
		if strings.Contains(p.Executable(), "sleep") {
			exists = true
		}
	}

	assert.False(t, exists)
	assert.Equal(t, app.status, Stopped)
}
