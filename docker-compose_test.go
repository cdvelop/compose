package compose

import (
	"testing"
)

func TestComposeGenerator(t *testing.T) {
	const testFile = "docker-compose.yml"

	// Función helper para crear una configuración de prueba
	createTestConfig := func() *ComposeConfig {
		return &ComposeConfig{
			Version: "3.8",
			Services: map[string]ServiceConfig{
				"api": {
					Image:         "golang:1.19",
					ContainerName: "api-service",
					Ports:         []string{"8080:8080"},
					Environment: map[string]string{
						"DB_HOST": "db",
						"DB_PORT": "5432",
					},
				},
				"db": {
					Image:         "postgres:14",
					ContainerName: "postgres-db",
					Environment: map[string]string{
						"POSTGRES_PASSWORD": "secretpassword",
						"POSTGRES_DB":       "myapp",
					},
				},
			},
		}
	}

	config := createTestConfig()
	err := config.SaveIfDifferent(testFile)
	if err != nil {
		t.Fatalf("Error inesperado: %v", err)
	}

}
