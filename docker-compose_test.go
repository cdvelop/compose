package compose_test

import (
	"os"
	"testing"

	"github.com/cdvelop/compose"

	"gopkg.in/yaml.v3"
)

func TestComposeGenerator(t *testing.T) {
	const testFile = "docker-compose.yml"

	dbService := *compose.NewService("db").
		SetContainerName("db").
		AddPort("5432", "5432").
		AddEnvironment("POSTGRES_DB", "POSTGRES_DB").
		AddEnvironment("POSTGRES_USER", "${POSTGRES_USER}").
		AddEnvironment("POSTGRES_PASSWORD", "${POSTGRES_PASSWORD}").
		SetImage("pgvector/pgvector:pg16").
		SetRestartPolicy("unless-stopped").
		AddVolume(compose.Volume{
			Source: "./init-db.sql",
			Target: "/docker-entrypoint-initdb.d/init-db.sql",
		}).
		SetHealthCheck(
			[]string{"CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ragtag"},
			"10s",
			"5s",
			5,
		)

	apiService := *compose.NewService("api").
		AddPort("8080", "8080").
		AddEnvironment("DB_HOST", "db").
		AddEnvironment("DB_PORT", "5432").
		SetImage("golang:1.19").
		DependsOn(dbService)

	config, err := compose.NewCompose("0.1", dbService, apiService)
	if err != nil {
		t.Fatalf("Error creando configuración: %v", err)
	}
	err = config.SaveIfDifferent(testFile)
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

	// Verificar container_name
	if api["container_name"] != "api" {
		t.Error("container_name incorrecto")
	}

	// Verificar dependencias
	dependencies, ok := api["depends_on"].([]interface{})
	if !ok || len(dependencies) != 1 || dependencies[0] != "db" {
		t.Error("Dependencias incorrectas")
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
