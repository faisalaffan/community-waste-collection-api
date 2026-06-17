package router

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/faisalaffan/community-waste-collection-api/docs"
	"github.com/faisalaffan/community-waste-collection-api/internal/handler"
)

func TestSetup_ReturnsNonNil(t *testing.T) {
	app := Setup(&handler.HouseholdHandler{}, &handler.PickupHandler{}, &handler.PaymentHandler{}, &handler.ReportHandler{}, &handler.FileHandler{})
	assert.NotNil(t, app)
}

func TestSetup_RoutesRegistered(t *testing.T) {
	app := Setup(&handler.HouseholdHandler{}, &handler.PickupHandler{}, &handler.PaymentHandler{}, &handler.ReportHandler{}, &handler.FileHandler{})
	routes := app.GetRoutes()
	assert.NotEmpty(t, routes)
}

func TestSetup_AllRoutes(t *testing.T) {
	// Create temporary web + assets files so SendFile succeeds
	tmp := t.TempDir()
	webDir := filepath.Join(tmp, "web")
	assetsDir := filepath.Join(tmp, "assets")
	require.NoError(t, os.MkdirAll(webDir, 0755))
	require.NoError(t, os.MkdirAll(assetsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(webDir, "index.html"), []byte("<html></html>"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(webDir, "app.js"), []byte("// app"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(assetsDir, "logo.png"), []byte("logo"), 0644))

	// Init swagger docs
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Title = "Test API"

	// Chdir to tmp so SendFile finds web/
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	app := Setup(&handler.HouseholdHandler{}, &handler.PickupHandler{}, &handler.PaymentHandler{}, &handler.ReportHandler{}, &handler.FileHandler{})

	// Test static + Swagger routes that don't need real services
	staticRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/app.js"},
		{"GET", "/assets/logo.png"},
		{"GET", "/swagger"},
		{"GET", "/swagger/doc.json"},
	}

	for _, tt := range staticRoutes {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)
		})
	}
}
