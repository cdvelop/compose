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
	name                string
	image               string
	containerName       string
	ports               []string
	environment         map[string]string
	volumes             []string
	serviceDependencies []string
	command             string
	networks            []string
}

// MarshalYAML implementa la interfaz yaml.Marshaler para Service
func (s Service) MarshalYAML() (interface{}, error) {
	return struct {
		Image         string            `yaml:"image"`
		ContainerName string            `yaml:"container_name,omitempty"`
		Ports         []string          `yaml:"ports,omitempty"`
		Environment   map[string]string `yaml:"environment,omitempty"`
		Volumes       []string          `yaml:"volumes,omitempty"`
		DependsOn     []string          `yaml:"depends_on,omitempty"`
		Command       string            `yaml:"command,omitempty"`
		Networks      []string          `yaml:"networks,omitempty"`
	}{
		Image:         s.image,
		ContainerName: s.containerName,
		Ports:         s.ports,
		Environment:   s.environment,
		Volumes:       s.volumes,
		DependsOn:     s.serviceDependencies,
		Command:       s.command,
		Networks:      s.networks,
	}, nil
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

// composeConfig representa la estructura completa del docker-compose
type composeConfig struct {
	version  string    `yaml:"version"`
	services []Service `yaml:"services"`
	volumes  []Volume  `yaml:"volumes,omitempty"`
	networks []Network `yaml:"networks,omitempty"`
}

// NewCompose crea una nueva configuración de docker-compose
func NewCompose(version string, services ...Service) (*composeConfig, error) {
	config := &composeConfig{
		version:  version,
		services: services,
	}

	// Extraer volúmenes y redes de los servicios
	volumesMap := make(map[string]Volume)
	networksMap := make(map[string]Network)

	for _, service := range services {
		// Extraer volúmenes
		for _, vol := range service.volumes {
			if vol != "" {
				volumesMap[vol] = Volume{
					Name:   vol,
					Driver: "local",
				}
			}
		}

		// Extraer redes
		for _, net := range service.networks {
			if net != "" {
				networksMap[net] = Network{
					Name:   net,
					Driver: "bridge",
				}
			}
		}
	}

	// Convertir maps a slices
	for _, vol := range volumesMap {
		config.volumes = append(config.volumes, vol)
	}

	for _, net := range networksMap {
		config.networks = append(config.networks, net)
	}

	return config, nil
}

// orderedMap es una estructura que preserva el orden de inserción
type orderedMap struct {
	keys   []string
	values map[string]interface{}
}

// MarshalYAML implementa la interfaz yaml.Marshaler para generar el formato correcto
func (c composeConfig) MarshalYAML() (interface{}, error) {
	// Crear map ordenado de servicios
	services := &orderedMap{
		keys:   make([]string, 0, len(c.services)),
		values: make(map[string]interface{}),
	}

	for _, s := range c.services {
		services.keys = append(services.keys, s.containerName)
		services.values[s.containerName] = s
	}

	// Crear map ordenado de volúmenes
	volumes := &orderedMap{
		keys:   make([]string, 0, len(c.volumes)),
		values: make(map[string]interface{}),
	}

	for _, v := range c.volumes {
		volumes.keys = append(volumes.keys, v.Name)
		volumes.values[v.Name] = v
	}

	// Crear estructura final
	type finalConfig struct {
		Version  string      `yaml:"version"`
		Services *orderedMap `yaml:"services"`
		Volumes  *orderedMap `yaml:"volumes,omitempty"`
		Networks *orderedMap `yaml:"networks,omitempty"`
	}

	config := finalConfig{
		Version:  c.version,
		Services: services,
		Volumes:  volumes,
	}

	// Solo agregar networks si hay redes definidas
	if len(c.networks) > 0 {
		// Crear map ordenado de redes
		networks := &orderedMap{
			keys:   make([]string, 0, len(c.networks)),
			values: make(map[string]interface{}),
		}

		for _, n := range c.networks {
			networks.keys = append(networks.keys, n.Name)
			networks.values[n.Name] = n
		}

		config.Networks = networks
	}

	return config, nil
}

// MarshalYAML implementa la interfaz yaml.Marshaler para orderedMap
func (m *orderedMap) MarshalYAML() (interface{}, error) {
	result := make(map[string]interface{})
	for _, key := range m.keys {
		result[key] = m.values[key]
	}
	return result, nil
}

// NewService crea una nueva configuración de servicio
func NewService(name string) *Service {
	return &Service{
		name:                name,
		containerName:       name,
		ports:               []string{},
		environment:         make(map[string]string),
		volumes:             []string{},
		serviceDependencies: []string{},
		networks:            []string{},
	}
}

// AddPort añade un mapeo de puertos al servicio
func (s *Service) AddPort(host, container string) *Service {
	s.ports = append(s.ports, fmt.Sprintf("%s:%s", host, container))
	return s
}

// AddEnvironment añade una variable de entorno al servicio
func (s *Service) AddEnvironment(key, value string) *Service {
	s.environment[key] = value
	return s
}

// AddVolume añade un volumen al servicio
func (s *Service) AddVolume(volume Volume) *Service {
	s.volumes = append(s.volumes, volume.Name)
	return s
}

// SetImage establece la imagen del servicio
func (s *Service) SetImage(image string) *Service {
	s.image = image
	return s
}

// DependsOn establece las dependencias del servicio
func (s *Service) DependsOn(services ...Service) *Service {
	for _, service := range services {
		s.serviceDependencies = append(s.serviceDependencies, service.name)
	}
	return s
}

// SaveIfDifferent guarda el archivo docker-compose.yml solo si es diferente del existente
func (c *composeConfig) SaveIfDifferent(filename ...string) error {
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
