package compose

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// PortMapping representa un mapeo de puertos
type PortMapping string

// Environment representa una variable de entorno
type Environment map[string]string

// NetworkConfig representa la configuración de una red
type NetworkConfig struct {
	Name    string            `yaml:"name,omitempty"`
	Driver  string            `yaml:"driver,omitempty"`
	Options map[string]string `yaml:"options,omitempty"`
}

// Service representa un servicio en docker-compose
type Service struct {
	Name          string            `yaml:"-"`
	Image         string            `yaml:"image"`
	ContainerName string            `yaml:"container_name,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
	DependsOn     []string          `yaml:"depends_on,omitempty"`
	Command       string            `yaml:"command,omitempty"`
	Networks      []string          `yaml:"networks,omitempty"`
}

// Volume representa un volumen en docker-compose
type Volume struct {
	Name   string            `yaml:"-"`
	Driver string            `yaml:"driver,omitempty"`
	Labels map[string]string `yaml:"labels,omitempty"`
}

// Network representa una red en docker-compose
type Network struct {
	Name    string            `yaml:"-"`
	Driver  string            `yaml:"driver,omitempty"`
	Options map[string]string `yaml:"options,omitempty"`
}

// ComposeConfig representa la estructura completa del docker-compose
type ComposeConfig struct {
	Version  string    `yaml:"version"`
	Services []Service `yaml:"services"`
	Volumes  []Volume  `yaml:"volumes,omitempty"`
	Networks []Network `yaml:"networks,omitempty"`
}

// MarshalYAML implementa la interfaz yaml.Marshaler para generar el formato correcto
func (c ComposeConfig) MarshalYAML() (interface{}, error) {
	// Convertir slices a maps para mantener compatibilidad con formato docker-compose
	servicesMap := make(map[string]Service)
	for _, s := range c.Services {
		servicesMap[s.Name] = s
	}

	volumesMap := make(map[string]Volume)
	for _, v := range c.Volumes {
		volumesMap[v.Name] = v
	}

	networksMap := make(map[string]Network)
	for _, n := range c.Networks {
		networksMap[n.Name] = n
	}

	return struct {
		Version  string             `yaml:"version"`
		Services map[string]Service `yaml:"services"`
		Volumes  map[string]Volume  `yaml:"volumes,omitempty"`
		Networks map[string]Network `yaml:"networks,omitempty"`
	}{
		Version:  c.Version,
		Services: servicesMap,
		Volumes:  volumesMap,
		Networks: networksMap,
	}, nil
}

// NewService crea una nueva configuración de servicio
func NewService(name string) *Service {
	return &Service{
		Name:          name,
		ContainerName: name,
		Ports:         []string{},
		Environment:   map[string]string{},
		Volumes:       []string{},
		DependsOn:     []string{},
		Networks:      []string{},
	}
}

// AddPort añade un mapeo de puertos al servicio
func (s *Service) AddPort(host, container string) *Service {
	s.Ports = append(s.Ports, fmt.Sprintf("%s:%s", host, container))
	return s
}

// AddEnvironment añade una variable de entorno al servicio
func (s *Service) AddEnvironment(key, value string) *Service {
	s.Environment[key] = value
	return s
}

// AddVolume añade un volumen al servicio
func (s *Service) AddVolume(source, destination, mode string) *Service {
	volume := source + ":" + destination
	if mode != "" {
		volume += ":" + mode
	}
	s.Volumes = append(s.Volumes, volume)
	return s
}

// SetImage establece la imagen del servicio
func (s *Service) SetImage(image string) *Service {
	s.Image = image
	return s
}

// SaveIfDifferent guarda el archivo docker-compose.yml solo si es diferente del existente
func (c *ComposeConfig) SaveIfDifferent(filename ...string) error {
	composePath := "docker-compose.yml"
	if len(filename) > 0 {
		composePath = filename[0]
	}

	// Generar nuevo YAML
	yamlData, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error al generar YAML: %v", err)
	}

	// Verificar si existe archivo actual
	currentData, err := os.ReadFile(composePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Si no existe, crear nuevo archivo
			return os.WriteFile(composePath, yamlData, 0644)
		}
		return fmt.Errorf("error al leer archivo: %v", err)
	}

	// Si el contenido es igual, no hacer nada
	if string(currentData) == string(yamlData) {
		return nil
	}

	// Guardar nuevo archivo si es diferente
	return os.WriteFile(composePath, yamlData, 0644)
}
