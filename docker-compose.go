package compose

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ServiceConfig representa la configuración de un servicio
type ServiceConfig struct {
	Image         string            `yaml:"image"`
	ContainerName string            `yaml:"container_name,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
	DependsOn     []string          `yaml:"depends_on,omitempty"`
	Command       string            `yaml:"command,omitempty"`
}

// ComposeConfig representa la estructura completa del docker-compose
type ComposeConfig struct {
	Version  string                   `yaml:"version"`
	Services map[string]ServiceConfig `yaml:"services"`
	Volumes  map[string]interface{}   `yaml:"volumes,omitempty"`
	Networks map[string]interface{}   `yaml:"networks,omitempty"`
}

// SaveIfDifferent guarda el archivo docker-compose.yml solo si es diferente del existente
func (c *ComposeConfig) SaveIfDifferent(filename string) error {
	// Generar nuevo YAML
	yamlData, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error al generar YAML: %v", err)
	}

	// Verificar si existe archivo actual
	currentData, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Si no existe, crear nuevo archivo
			return os.WriteFile(filename, yamlData, 0644)
		}
		return fmt.Errorf("error al leer archivo: %v", err)
	}

	// Si el contenido es igual, no hacer nada
	if string(currentData) == string(yamlData) {
		return nil
	}

	// Guardar nuevo archivo si es diferente
	return os.WriteFile(filename, yamlData, 0644)
}
