package compose

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// healthCheck representa la configuración de healthcheck
type healthCheck struct {
	Test     []string
	Interval string
	Timeout  string
	Retries  int
}

// service representa un servicio en docker-compose
type service struct {
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
	healthCheck         *healthCheck
	errors              []error
}

// SetRestartPolicy establece la política de reinicio del servicio
func (s *service) SetRestartPolicy(policy string) *service {
	s.restartPolicy = policy
	return s
}

// SetHealthCheck configura el healthcheck del servicio
func (s *service) SetHealthCheck(test []string, interval, timeout string, retries int) *service {
	s.healthCheck = &healthCheck{
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

// composeConfig representa la estructura completa del docker-compose
type composeConfig struct {
	version  string    `yaml:"version"`
	services []service `yaml:"services"`
	volumes  []Volume  `yaml:"volumes,omitempty"`
}

// NewCompose crea una nueva configuración de docker-compose
func NewCompose(version string, services ...service) (*composeConfig, error) {
	config := &composeConfig{
		version:  version,
		services: services,
	}

	return config, nil
}

// generateYAML genera el contenido YAML respetando el orden de los servicios
func (c composeConfig) generateYAML() ([]byte, error) {
	var b strings.Builder

	var out_errors []error
	// Escribir versión
	fmt.Fprintf(&b, "version: %q\n", c.version)

	// Escribir servicios
	b.WriteString("services:\n")
	for _, service := range c.services {

		if len(service.errors) > 0 {
			out_errors = append(out_errors, service.errors...)
			continue
		}

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

	if len(out_errors) > 0 {
		return nil, errors.Join(out_errors...)
	}

	return []byte(b.String()), nil
}

// NewService crea una nueva configuración de servicio
func NewService(name string) *service {
	return &service{
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
func (s *service) SetContainerName(name string) *service {
	s.containerName = name
	return s
}

// AddPort añade un mapeo de puertos al servicio
func (s *service) AddPort(host, container string) *service {
	s.ports = append(s.ports, fmt.Sprintf("%s:%s", host, container))
	return s
}

// AddEnvironment adds an environment variable to the service
// If a value is provided, it will be used for both public and private values
// If no value is provided, it will look for the variable in the environment
// and use ${key} for the public value and the actual value for the private value
// The private value will be added to the .env file
func (s *service) AddEnvironment(key string, value ...string) *service {
	var envPubValue, envPrivValue string

	if len(value) > 0 {
		envPubValue = value[0]
		envPrivValue = value[0]
	} else {
		// Buscar en variables de entorno
		val, exists := os.LookupEnv(key)
		if !exists {
			s.errors = append(s.errors, fmt.Errorf("environment variable %s not found", key))
			return s
		}
		// Usar ${key} para el valor público
		envPubValue = fmt.Sprintf("${%s}", key)
		// Usar el valor real para el privado
		envPrivValue = val
	}

	if envPrivValue != "" {
		AddEnvToFile(key, envPrivValue)
	}

	s.environment[key] = envPubValue
	return s
}

// AddVolume añade un volumen al servicio
func (s *service) AddVolume(volume Volume) *service {
	s.volumes = append(s.volumes, volume)
	return s
}

// SetImage establece la imagen del servicio
func (s *service) SetImage(image string) *service {
	s.image = image
	return s
}

// DependsOn establece las dependencias del servicio
func (s *service) DependsOn(services ...service) *service {
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
