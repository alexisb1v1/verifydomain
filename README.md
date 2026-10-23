# verifyDomain - Microservicio Transversal de Autorización Perimetral TLS

> **Componente de Infraestructura Perimetral**  
> Lenguaje: **Go 1.22+**  
> Rol: Verificación ultrarrápida de subdominios multi-tenant para emisión On-Demand TLS en **Caddy Proxy**.

---

## 1. Propósito y Arquitectura

`verifyDomain` es un microservicio perimetral de alto rendimiento diseñado para integrarse con la directriz `on_demand_tls` de **Caddy**.  
Antes de que Caddy solicite o renueve un certificado SSL (Let's Encrypt / ZeroSSL) para un subdominio entrante (ej. `empresa.neocobros.com` o `flota.neoruta.com`), consulta a `verifyDomain` vía HTTP para validar si el dominio está registrado y activo.

```
                              ┌────────────────────────────────────────┐
                              │           Caddy Reverse Proxy          │
                              └───────┬────────────────────────┬───────┘
                                      │                        │
       (A) Subdominios Dinámicos      │                        │ (B) Dominios Fijos
       https://*.neocobros.com        │                        │ https://central.neocobros.com
       [on_demand_tls]                │                        │ [Registro Fijo en Caddyfile]
                                      ▼                        ▼
                       ┌─────────────────────────────┐   ┌───────────────────────────┐
                       │        verifyDomain         │   │ Certificado Directo Caddy │
                       └──────┬───────────────┬──────┘   │ (Bypass verifyDomain y BD)│
                              │               │          └───────────────────────────┘
       1. GET tenant:<domain> │               │ 2. Fallback (Self-Healing)
                              ▼               ▼
                       ┌─────────────┐  ┌──────────────────┐
                       │ Redis Cache │  │ Backend APIs     │
                       │(tenant:...) │  │ - PrestamosApi   │
                       └─────────────┘  │ - VecturaApi     │
                                        └──────────────────┘
```

> [!NOTE]
> **Dominio Fijo `central.neocobros.com` (Exento de verifyDomain):**  
> El dominio maestro `https://central.neocobros.com` tiene una directiva estática fija en el `Caddyfile` y maneja su propio certificado SSL. **No utiliza `on_demand_tls` ni pasa por este microservicio**. En consecuencia, `verifyDomain` solo procesa peticiones de tenants clientes (`<empresa>.neocobros.com`), y `PrestamosApi` solo resuelve empresas clientes registradas en su base de datos.

---

## 2. Contrato Inmutable de Redis (`tenant:<domain>`)

> [!CAUTION]
> **REGLA DE ORO ARQUITECTÓNICA - NO MODIFICAR ESTE FORMATO DE CLAVE:**  
> La clave en Redis **`tenant:<domain>`** es un contrato estricto compartido transversalmente entre múltiples repositorios:
> - `verifyDomain` (este servicio en Go)
> - `PrestamosApi` (NeoCobros)
> - `VecturaApi` (NeoRuta)
> - `Caddy`
> 
> **NUNCA agregarle prefijos privados** (como `neocobros:tenant:`). Hacerlo desconectará inmediatamente la resolución TLS de todos los subdominios y provocará fallos 502 / SSL Handshake Error en producción.

### Formato de Clave:
```
tenant:<subdominio>.<dominio_base>
```
* **Ejemplos válidos:**
  * `tenant:demo.neocobros.com` -> `"1"`
  * `tenant:central.neoruta.com` -> `"1"`
* **Valor:** `"1"` indica que el dominio está activo y autorizado.
* **TTL:** Persistente (gestionado por los comandos de backend al crear, editar o desactivar empresas).

---

## 3. Mecanismo de Auto-Curación (*Self-Healing Fallback*)

Si la clave en Redis no se encuentra (por ejemplo, tras un reinicio de Redis o antes de una sincronización masiva), `verifyDomain` no rechaza la petición de inmediato.  
En su lugar, ejecuta una llamada de auto-curación (*Self-Healing*) hacia el backend responsable basándose en el sufijo del dominio:

| Sufijo del Dominio | Variable de Entorno API | Endpoint de Fallback Consultado |
| :--- | :--- | :--- |
| `*.neocobros.com` | `PRESTAMOS_API_URL` | `GET /api/v1/company/verify-domain?domain=<domain>` |
| `*.neoruta.com` | `VECTURA_API_URL` | `GET /v1/tenants/verify-domain?domain=<domain>` |

Si la API del backend responde con `200 OK`, el backend mismo restaura la clave `tenant:<domain> = "1"` en Redis y `verifyDomain` autoriza a Caddy para emitir el certificado sin interrupción de servicio.

---

## 4. Endpoints HTTP

### `GET /api/v1/tenants/verify-domain?domain={subdominio.dominio.com}`

* **Query Params:**
  * `domain` (requerido): El FQDN que Caddy está intentando validar (ej. `empresa.neocobros.com`).
* **Respuestas:**
  * `200 OK`: Dominio autorizado. Caddy procede con el handshake TLS.
    ```json
    { "verified": true, "domain": "empresa.neocobros.com" }
    ```
  * `403 Forbidden`: Dominio inactivo, deshabilitado o no registrado.
    ```json
    { "verified": false, "error": "Dominio no registrado o inactivo" }
    ```
  * `400 Bad Request`: Falta el parámetro `domain`.
  * `405 Method Not Allowed`: Método HTTP distinto a `GET`.
  * `500 Internal Server Error`: Falla de conectividad crítica con Redis.

---

## 5. Variables de Entorno (`.env`)

```ini
PORT=3000
REDIS_HOST=127.0.0.1
REDIS_PORT=6380
REDIS_PASSWORD=
REDIS_DB=0

# Configuración NeoCobros
PRESTAMOS_BASE_DOMAIN=neocobros.com
PRESTAMOS_API_URL=http://prestamos-api:3000

# Configuración NeoRuta / Vectura
VECTURA_BASE_DOMAIN=neoruta.com
VECTURA_API_URL=http://vectura-api:3000
```
