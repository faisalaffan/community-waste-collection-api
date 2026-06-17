package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetup_ReturnsNonNil(t *testing.T) {
	app := Setup(nil, nil, nil, nil, nil)
	assert.NotNil(t, app)
}

func TestSetup_RoutesRegistered(t *testing.T) {
	app := Setup(nil, nil, nil, nil, nil)
	assert.NotNil(t, app)

	routes := app.GetRoutes()
	assert.NotEmpty(t, routes)
}
