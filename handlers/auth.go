package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/redis/go-redis/v9"
	"verifydomain/config"
	"verifydomain/db"
)

// VerifyResponse define la estructura JSON de la respuesta de la API
type VerifyResponse struct {
	Verified bool   `json:"verified"`
	Domain   string `json:"domain,omitempty"`
	Error    string `json:"error,omitempty"`
}

func checkApiFallback(domain string) bool {
	cfg := config.GlobalConfig
	var apiUrl string

	if cfg.PrestamosBaseDomain != "" && strings.HasSuffix(domain, cfg.PrestamosBaseDomain) {
		apiUrl = fmt.Sprintf("%s/api/v1/company/verify-domain?domain=%s", cfg.PrestamosApiUrl, domain)
	} else if cfg.VecturaBaseDomain != "" && strings.HasSuffix(domain, cfg.VecturaBaseDomain) {
		apiUrl = fmt.Sprintf("%s/v1/tenants/verify-domain?domain=%s", cfg.VecturaApiUrl, domain)
	} else {
		// No coincide con ninguno de los dominios base
		return false
	}

	resp, err := http.Get(apiUrl)
	if err != nil {
		log.Printf("Error al consultar API de fallback (%s): %v", apiUrl, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("[Self-Healing] La API (%s) autorizó el dominio %s y regeneró la caché.", apiUrl, domain)
		return true
	}

	return false
}

// VerifyDomainHandler maneja la petición de verificación del dominio para Caddy.
// Busca la clave 'tenant:<domain>' en Redis y responde con 200 OK si existe y su valor es "1".
func VerifyDomainHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Validar que la petición sea de tipo GET
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(VerifyResponse{
			Verified: false,
			Error:    "Método no permitido. Utilice GET",
		})
		return
	}

	// 2. Extraer el parámetro query 'domain'
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(VerifyResponse{
			Verified: false,
			Error:    "El parámetro 'domain' es requerido",
		})
		return
	}

	// 3. Consultar la existencia del dominio en Redis
	val, err := db.GetTenantValue(r.Context(), domain)
	if err != nil {
		// Caso: La clave no existe en Redis
		if errors.Is(err, redis.Nil) {
			// Intentar Fallback a la API correspondiente según el subfijo
			if checkApiFallback(domain) {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(VerifyResponse{
					Verified: true,
					Domain:   domain,
				})
				return
			}

			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(VerifyResponse{
				Verified: false,
				Error:    "Dominio no registrado o inactivo",
			})
			return
		}

		// Caso: Error técnico con la conexión o comando en Redis
		log.Printf("Error de base de datos para el dominio [%s]: %v", domain, err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(VerifyResponse{
			Verified: false,
			Error:    "Error interno al verificar el dominio",
		})
		return
	}

	// 4. Validar el valor de la clave. Debe ser "1" para autorizar.
	if val == "1" {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(VerifyResponse{
			Verified: true,
			Domain:   domain,
		})
	} else {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(VerifyResponse{
			Verified: false,
			Error:    "Dominio inactivo o deshabilitado",
		})
	}
}
