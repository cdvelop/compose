package compose

import (
	"fmt"
	"os"
	"strings"
)

// NetworkConfig representa la configuración de una red
type etworkConfig struct {
	Name    string
	Driver  string
	Options map[string]string
}

// HealthCheck representa la configuración de healthcheck
type HealthCheck struct {
	Test     []string
	Interval string
	Timeout  string
	Retries  int
}

// Service representa un servicio en docker-compose
type Service struct {
	name                string
	image               string
	containerName       string
	ports               []string
	environment         map[string]string
	volumes             []Volume
	serviceDependencies []string
	command             string
	networks            []string
	restartPolicy       string
	healthCheck         *HealthCheck
}

// SetRestartPolicy establece la política de reinicio del servicio
func (s *Service) SetRestartPolicy(policy string) *Service {
	s.restartPolicy = policy
	return s
}

// SetHealthCheck configura el healthcheck del servicio
func (s *Service) SetHealthCheck(test []string, interval, timeout string, retries int) *Service {
	s.healthCheck = &HealthCheck{
		Test:     test,
		Interval: interval,
		Timeout:  timeout,
		Retries:  retries,
	}
	return s
}

// Volume representa un volumen en docker-compose
type Volume struct {
	Source string `yaml:"-"`
	Target string `yaml:"-"`
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

	// Extraer redes de los servicios
	networksMap := make(map[string]Network)

	for _, service := range services {
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

	// Convertir map de redes a slice
	for _, net := range networksMap {
		config.networks = append(config.networks, net)
	}

	return config, nil
}

// generateYAML genera el contenido YAML respetando el orden de los servicios
func (c composeConfig) generateYAML() ([]byte, error) {
	var b strings.Builder

	// Escribir versión
	fmt.Fprintf(&b, "version: %q\n", c.version)

	// Escribir servicios
	b.WriteString("services:\n")
	for _, service := range c.services {
		fmt.Fprintf(&b, "  %s:\n", service.containerName)
		fmt.Fprintf(&b, "    image: %q\n", service.image)

		if service.containerName != "" {
			fmt.Fprintf(&b, "    container_name: %q\n", service.containerName)
		}

		if len(service.ports) > 0 {
			b.WriteString("    ports:\n")
			for _, port := range service.ports {
				fmt.Fprintf(&b, "      - \"%s\"\n", port)
			}
		}

		if len(service.environment) > 0 {
			b.WriteString("    environment:\n")
			for key, value := range service.environment {
				fmt.Fprintf(&b, "      %q: %q\n", key, value)
			}
		}

		if len(service.volumes) > 0 {
			b.WriteString("    volumes:\n")
			for _, vol := range service.volumes {
				fmt.Fprintf(&b, "      - %s:%s\n", vol.Source, vol.Target)
			}
		}

		if len(service.serviceDependencies) > 0 {
			b.WriteString("    depends_on:\n")
			for _, dep := range service.serviceDependencies {
				fmt.Fprintf(&b, "      - %q\n", dep)
			}
		}

		if service.command != "" {
			fmt.Fprintf(&b, "    command: %q\n", service.command)
		}

		if len(service.networks) > 0 {
			b.WriteString("    networks:\n")
			for _, net := range service.networks {
				fmt.Fprintf(&b, "      - %q\n", net)
			}
		}

		if service.restartPolicy != "" {
			fmt.Fprintf(&b, "    restart: %q\n", service.restartPolicy)
		}

		if service.healthCheck != nil {
			b.WriteString("    healthcheck:\n")
			fmt.Fprintf(&b, "      test:\n")
			for _, test := range service.healthCheck.Test {
				fmt.Fprintf(&b, "        - %q\n", test)
			}
			if service.healthCheck.Interval != "" {
				fmt.Fprintf(&b, "      interval: %q\n", service.healthCheck.Interval)
			}
			if service.healthCheck.Timeout != "" {
				fmt.Fprintf(&b, "      timeout: %q\n", service.healthCheck.Timeout)
			}
			if service.healthCheck.Retries > 0 {
				fmt.Fprintf(&b, "      retries: %d\n", service.healthCheck.Retries)
			}
		}
	}

	// Escribir redes si existen
	if len(c.networks) > 0 {
		b.WriteString("networks:\n")
		for _, network := range c.networks {
			fmt.Fprintf(&b, "  %s:\n", network.Name)
			if network.Driver != "" {
				fmt.Fprintf(&b, "    driver: %q\n", network.Driver)
			}
			if len(network.Options) > 0 {
				b.WriteString("    options:\n")
				for key, value := range network.Options {
					fmt.Fprintf(&b, "      %q: %q\n", key, value)
				}
			}
		}
	}

	return []byte(b.String()), nil
}

// NewService crea una nueva configuración de servicio
func NewService(name string) *Service {
	return &Service{
		name:                name,
		containerName:       name,
		ports:               []string{},
		environment:         make(map[string]string),
		volumes:             []Volume{},
		serviceDependencies: []string{},
		networks:            []string{},
	}
}

// SetContainerName establece el nombre del contenedor
func (s *Service) SetContainerName(name string) *Service {
	s.containerName = name
	return s
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
	s.volumes = append(s.volumes, volume)
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

	// Generar nuevo YAML usando nuestra implementación personalizada
	yamlData, err := c.generateYAML()
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
