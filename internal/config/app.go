package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// AppConfig centraliza as variáveis de ambiente necessárias para configurar a aplicação,
// incluindo portas dos servidores REST e gRPC, e caminho do arquivo de persistência de estado.
type AppConfig struct {
	RestPort   string
	GrpcPort   string
	BackupFile string
}

// LoadConfig carrega variáveis de ambiente a partir de arquivo .env ou do sistema operacional,
// inicializando valores padrão para portas e caminho de backup quando não configurados.
func LoadConfig() *AppConfig {
	// Tenta carregar .env; em ambientes containerizados (Kubernetes), as variáveis virão do SO
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️ .env não encontrado. Utilizando variáveis de ambiente do SO.")
	}

	backupFile := os.Getenv("BACKUP_FILE")
	if backupFile == "" {
		backupFile = filepath.Join("build", "backup.json")
	}

	restPort := os.Getenv("REST_PORT")
	if restPort == "" {
		restPort = "8081"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}

	return &AppConfig{
		RestPort:   restPort,
		GrpcPort:   grpcPort,
		BackupFile: backupFile,
	}
}
