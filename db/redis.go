package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"verifydomain/config"
)

// Client es la instancia global expuesta para interactuar con Redis
var Client *redis.Client

// InitRedis inicializa el cliente de Redis y realiza un PING para comprobar la disponibilidad del servicio
func InitRedis(cfg *config.Config) error {
	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)

	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Timeout corto para validar que la conexión esté activa
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("falló el PING a Redis en %s: %w", addr, err)
	}

	return nil
}

// GetTenantValue busca la clave 'tenant:<domain>' en Redis y retorna su valor.
// Si la clave no existe, retornará el error redis.Nil.
func GetTenantValue(ctx context.Context, domain string) (string, error) {
	key := fmt.Sprintf("tenant:%s", domain)
	return Client.Get(ctx, key).Result()
}

