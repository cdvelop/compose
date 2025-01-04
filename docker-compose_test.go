package compose

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestComposeGenerator(t *testing.T) {
	const testFile = "docker-compose.yml"

	// Función helper para crear una configuración de prueba
	createTestConfig := func() *ComposeConfig {
		return &ComposeConfig{
			Version: "3.8",
			Services: []Service{
				*NewService("api").
					AddPort("8080", "8080").
					AddEnvironment("DB_HOST", "db").
					AddEnvironment("DB_PORT", "5432").
					SetImage("golang:1.19"),
				*NewService("db").
					AddEnvironment("POSTGRES_PASSWORD", "secretpassword").
					AddEnvironment("POSTGRES_DB", "myapp").
					SetImage("postgres:14"),
			},
		}
	}

	config := createTestConfig()
	err := config.SaveIfDifferent(testFile)
	if err != nil {
		t.Fatalf("Error inesperado: %v", err)
	}

	// Verificar el archivo YAML generado
	verifyGeneratedYAML(t, testFile)
}

func verifyGeneratedYAML(t *testing.T, filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Error leyendo archivo YAML: %v", err)
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("Error parseando YAML: %v", err)
	}

	// Verificar estructura básica
	if _, ok := result["version"]; !ok {
		t.Error("Falta campo 'version' en YAML")
	}

	services, ok := result["services"].(map[string]interface{})
	if !ok {
		t.Fatal("Estructura de servicios inválida")
	}

	if len(services) != 2 {
		t.Fatal("Número incorrecto de servicios")
	}

	// Verificar servicio API
	api, ok := services["api"].(map[string]interface{})
	if !ok {
		t.Fatal("Estructura de servicio API inválida")
	}

	// Verificar puertos
	ports, ok := api["ports"].([]interface{})
	if !ok || len(ports) != 1 || ports[0] != "8080:8080" {
		t.Error("Mapeo de puertos incorrecto")
	}

	// Verificar variables de entorno
	env, ok := api["environment"].(map[string]interface{})
	if !ok || len(env) != 2 {
		t.Error("Variables de entorno incorrectas")
	}
	if env["DB_HOST"] != "db" || env["DB_PORT"] != "5432" {
		t.Error("Valores de entorno incorrectos")
	}

	// Verificar que no hay campos vacíos
	for _, service := range services {
		s := service.(map[string]interface{})
		for key, value := range s {
			if value == nil || value == "" {
				t.Errorf("Campo vacío encontrado: %s", key)
			}
		}
	}
}
