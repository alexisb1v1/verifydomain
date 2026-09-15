package config

import (
	"log"
	"os"
	"strconv"
)

// Config almacena todas las variables de entorno necesarias para la aplicación
type Config struct {
	Port               string
	RedisHost          string
	RedisPort          string
	RedisPassword      string
	RedisDB            int
	PrestamosBaseDomain string
	PrestamosApiUrl     string
	VecturaBaseDomain   string
	VecturaApiUrl       string
}

// GlobalConfig almacena la configuración cargada de forma global para usarla desde otros paquetes
var GlobalConfig *Config

// LoadConfig carga la configuración desde las variables de entorno del sistema
func LoadConfig() *Config {
	port := getEnv("PORT", "3000")
	// Asegurar que el puerto tenga el prefijo ":" para http.ListenAndServe
	if port != "" && port[0] != ':' {
		port = ":" + port
	}

	redisDBStr := getEnv("REDIS_DB", "0")
	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		log.Printf("Advertencia: Valor de REDIS_DB inválido '%s', utilizando 0 por defecto", redisDBStr)
		redisDB = 0
	}

	GlobalConfig = &Config{
		Port:                port,
		RedisHost:           getEnv("REDIS_HOST", "127.0.0.1"),
		RedisPort:           getEnv("REDIS_PORT", "6380"),
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),
		RedisDB:             redisDB,
		PrestamosBaseDomain: getEnv("PRESTAMOS_BASE_DOMAIN", "neocobros.com"),
		PrestamosApiUrl:     getEnv("PRESTAMOS_API_URL", "http://prestamos-api:3000"),
		VecturaBaseDomain:   getEnv("VECTURA_BASE_DOMAIN", "neoruta.com"),
		VecturaApiUrl:       getEnv("VECTURA_API_URL", "http://vectura-api:3000"),
	}
	return GlobalConfig
}

// getEnv busca una variable de entorno y retorna un valor por defecto si no está establecida
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
