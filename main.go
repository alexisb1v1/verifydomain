package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"verifydomain/config"
	"verifydomain/db"
	"verifydomain/handlers"
)

func main() {
	// 1. Cargar archivo .env (silencioso si no existe, ya que en producción se inyectan variables directamente)
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: No se pudo cargar el archivo .env (se usarán variables de entorno del sistema)")
	}

	// 2. Cargar la configuración
	cfg := config.LoadConfig()

	// 3. Inicializar la conexión a Redis
	if err := db.InitRedis(cfg); err != nil {
		log.Fatalf("Error crítico al inicializar la base de datos Redis: %v", err)
	}
	log.Printf("Conectado exitosamente a Redis en %s:%s (DB: %d)", cfg.RedisHost, cfg.RedisPort, cfg.RedisDB)

	// 4. Registrar enrutador HTTP
	http.HandleFunc("/api/v1/tenants/verify-domain", handlers.VerifyDomainHandler)

	// 5. Levantar el servidor HTTP
	log.Printf("Servidor de verificación escuchando en el puerto %s", cfg.Port)
	if err := http.ListenAndServe(cfg.Port, nil); err != nil {
		log.Fatalf("Error crítico al iniciar el servidor HTTP: %v", err)
	}
}
