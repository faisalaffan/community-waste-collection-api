package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetup_ReturnsNonNil(t *testing.T) {
	app := Setup(nil, nil, nil, nil)
	assert.NotNil(t, app)
}

func TestSetup_RoutesRegistered(t *testing.T) {
	app := Setup(nil, nil, nil, nil)
	assert.NotNil(t, app)

	// Verify the app has routes by checking it handles requests without panic
	// (nil handlers will cause 500 but routes exist)
	routes := app.GetRoutes()
	assert.NotEmpty(t, routes)
}
