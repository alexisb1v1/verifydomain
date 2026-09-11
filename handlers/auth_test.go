package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestVerifyDomainHandler_MissingDomain comprueba que el endpoint responda con 400 Bad Request
// cuando no se envía el parámetro de consulta 'domain'.
func TestVerifyDomainHandler_MissingDomain(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/tenants/verify-domain", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(VerifyDomainHandler)

	handler.ServeHTTP(rr, req)

	// Verificar código de estado HTTP
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Se esperaba estado %v, pero se obtuvo %v", http.StatusBadRequest, rr.Code)
	}

	// Verificar cuerpo JSON de respuesta
	var resp VerifyResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("No se pudo decodificar el cuerpo JSON: %v", err)
	}

	if resp.Verified {
		t.Error("Se esperaba que 'verified' fuera falso")
	}

	const expectedError = "El parámetro 'domain' es requerido"
	if resp.Error != expectedError {
		t.Errorf("Se esperaba el error %q, pero se obtuvo %q", expectedError, resp.Error)
	}
}

// TestVerifyDomainHandler_InvalidMethod comprueba que el endpoint responda con 405 Method Not Allowed
// ante métodos distintos a GET.
func TestVerifyDomainHandler_InvalidMethod(t *testing.T) {
	req, err := http.NewRequest("POST", "/api/v1/tenants/verify-domain?domain=test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(VerifyDomainHandler)

	handler.ServeHTTP(rr, req)

	// Verificar código de estado HTTP
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Se esperaba estado %v, pero se obtuvo %v", http.StatusMethodNotAllowed, rr.Code)
	}

	// Verificar cuerpo JSON de respuesta
	var resp VerifyResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("No se pudo decodificar el cuerpo JSON: %v", err)
	}

	if resp.Verified {
		t.Error("Se esperaba que 'verified' fuera falso")
	}
}
