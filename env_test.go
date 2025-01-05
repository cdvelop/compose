package compose_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cdvelop/compose"
)

func TestAddEnvToFile(t *testing.T) {
	testDir := "testEnv"
	envPath := filepath.Join(testDir, ".env")
	gitignorePath := filepath.Join(testDir, ".gitignore")

	// Limpiar archivos antes de cada test
	cleanupFiles := func() {
		os.Remove(envPath)
		os.Remove(gitignorePath)
	}

	t.Run("Crear nuevo .env cuando no existe", func(t *testing.T) {
		defer cleanupFiles()

		err := compose.AddEnvToFile("DB_HOST", "localhost", envPath, gitignorePath)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}

		// Verificar que .env fue creado
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			t.Error(".env no fue creado")
		}

		// Verificar contenido
		content, err := os.ReadFile(envPath)
		if err != nil {
			t.Fatalf("Error leyendo .env: %v", err)
		}

		expected := "DB_HOST=localhost\n"
		if string(content) != expected {
			t.Errorf("Contenido inesperado:\nEsperado: %q\nObtenido: %q", expected, string(content))
		}

		// Verificar .gitignore
		gitignore, err := os.ReadFile(gitignorePath)
		if err != nil {
			t.Fatalf("Error leyendo .gitignore: %v", err)
		}

		if !strings.Contains(string(gitignore), ".env") {
			t.Error(".env no fue agregado a .gitignore")
		}
	})

	t.Run("Agregar nueva variable a .env existente", func(t *testing.T) {
		defer cleanupFiles()

		// Crear .env inicial
		err := os.WriteFile(envPath, []byte("EXISTING=value\n"), 0644)
		if err != nil {
			t.Fatalf("Error creando .env inicial: %v", err)
		}

		err = compose.AddEnvToFile("NEW_VAR", "new_value", envPath, gitignorePath)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}

		content, err := os.ReadFile(envPath)
		if err != nil {
			t.Fatalf("Error leyendo .env: %v", err)
		}

		expected := "EXISTING=value\nNEW_VAR=new_value\n"
		if string(content) != expected {
			t.Errorf("Contenido inesperado:\nEsperado: %q\nObtenido: %q", expected, string(content))
		}
	})

	t.Run("Actualizar variable existente", func(t *testing.T) {
		defer cleanupFiles()

		// Crear .env inicial
		err := os.WriteFile(envPath, []byte("EXISTING=old_value\n"), 0644)
		if err != nil {
			t.Fatalf("Error creando .env inicial: %v", err)
		}

		err = compose.AddEnvToFile("EXISTING", "new_value", envPath, gitignorePath)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}

		content, err := os.ReadFile(envPath)
		if err != nil {
			t.Fatalf("Error leyendo .env: %v", err)
		}

		expected := "EXISTING=new_value\n"
		if string(content) != expected {
			t.Errorf("Contenido inesperado:\nEsperado: %q\nObtenido: %q", expected, string(content))
		}
	})

	t.Run("Manejo correcto de .gitignore", func(t *testing.T) {
		defer cleanupFiles()

		// Crear .gitignore inicial
		err := os.WriteFile(gitignorePath, []byte("*.log\n"), 0644)
		if err != nil {
			t.Fatalf("Error creando .gitignore inicial: %v", err)
		}

		err = compose.AddEnvToFile("TEST_VAR", "value", envPath, gitignorePath)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}

		content, err := os.ReadFile(gitignorePath)
		if err != nil {
			t.Fatalf("Error leyendo .gitignore: %v", err)
		}

		expected := "*.log\n.env\n"
		if string(content) != expected {
			t.Errorf("Contenido inesperado de .gitignore:\nEsperado: %q\nObtenido: %q", expected, string(content))
		}
	})
}
