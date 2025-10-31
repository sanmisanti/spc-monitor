# Sistema de Monitoreo de Infraestructura - Secretaría de Contrataciones

## Contexto del Proyecto

Este proyecto surge de la necesidad de monitorear múltiples sistemas desplegados en diferentes ambientes para la Secretaría de Contrataciones:

- Servidores propios (on-premise)
- Servicios en la nube
- Entornos de Google Cloud Platform
- Otros ambientes heterogéneos

## Objetivo Principal

Desarrollar una aplicación web centralizada que permita monitorear el estado operacional de todos los sistemas bajo responsabilidad del área de ingeniería informática.

## Funcionalidades Identificadas

### 1. Monitoreo de Disponibilidad
- Verificar si los sistemas están online
- Health checks de servicios y aplicaciones

### 2. Monitoreo de Bases de Datos
- Verificar que se estén registrando datos (actividad de escritura)
- Consultar el estado de registros específicos
- Detectar anomalías en la actividad de BD

### 3. Integración con DevOps
- Revisar mensajes sin responder en plataformas DevOps
- Monitorear pipelines y deployments
- Alertas de builds fallidos

### 4. Funcionalidades Adicionales
- Por descubrir durante el desarrollo iterativo

## Metodología de Trabajo

### Principios de Desarrollo

1. **Análisis Fundamentado**: Cada decisión técnica debe estar respaldada por análisis y justificación objetiva

2. **Iteración y Validación**: No tomar decisiones definitivas sin explorar alternativas y validar supuestos

3. **Precisión Técnica**: Toda información proporcionada debe ser verificable y precisa

4. **Colaboración Activa**: Proceso de ida y vuelta para encontrar las mejores soluciones

5. **Sin Condescendencia**: Comunicación directa, técnica y profesional

6. **Decisiones Informadas**: No inventar soluciones ni tomar atajos sin fundamento

## Sistemas a Monitorear

### 1. SaltaCompra (Plataforma de Contrataciones Provincial)
- **Tecnología**: ASP.NET
- **Infraestructura**: Servidores propios Windows Server
- **Base de Datos**: SQL Server
- **Ambientes**:
  - **Producción**:
    - VM dedicada para servidor web (Windows Server)
    - VM dedicada para base de datos SQL Server
    - URL: `https://saltacompra.gob.ar/`
  - **Preproducción**:
    - VM dedicada para servidor web (Windows Server)
    - VM dedicada para base de datos SQL Server
    - URL: `https://preproduccion.saltacompra.gob.ar/`

### 2. App.SaltaCompra
- **Tecnología**: Next.js + PM2
- **Infraestructura**: VM Ubuntu dedicada (IP: 172.17.1.241)
- **Base de Datos**: PostgreSQL 16.2
  - Base de datos: `app_spc`
  - Schema principal: `users`
  - Tabla monitoreada: `users.usuarios`
- **URL**: `https://app.saltacompra.gob.ar/`
- **Ubicación**: Dentro de VPN (requiere conexión VPN para checks de BD)
- **Checks implementados**:
  - HTTP: Accesible públicamente (con certificado SSL autofirmado/CA privada)
  - PostgreSQL: Requiere VPN, verifica conteo de usuarios
- **Nota**: Servidor contiene web + base de datos en la misma VM

### 3. Sistema Google Apps Script - Kairos
- **Tecnología**: Google Apps Script
- **Infraestructura**: Google Cloud Platform
- **Storage**: Google Spreadsheets (como BD relacional)
- **Funcionalidad**: Scripts de automatización para actualización de base de datos de precios Kairos
- **Hoja monitoreada**: "Log Actualizaciones Kairos"
  - Registro diario automático de actualizaciones
  - Columnas monitoreadas:
    - TimeStamp (columna A): Fecha en formato YYYY-MM-DD
    - Nombre Archivo (columna C): Base de datos utilizada (bd_YYYYMMDD.zip)
  - **Validación**: Verificar que la última actualización sea del día actual y use la BD correspondiente al mismo día

### 4. Infraestructura Compartida
- **Dominio**: `saltacompra.gob.ar` (compartido por prod, preprod y app)
- **Monitoreo**: Expiración de dominio vía RDAP (NIC Argentina)
- **Criticidad**: Alta (afecta a todos los sistemas si expira)

**Nota**: Sistema extensible para agregar más aplicaciones en el futuro.

## Decisiones Técnicas

### Stack Tecnológico Seleccionado

**Decisión tomada tras análisis de requisitos:**
- Aplicación interna (sin necesidad de SEO)
- Prioridad: rendimiento, eficiencia de recursos, UI moderna
- Restricción: servidor compartido con otras aplicaciones
- Requisito: 100% open source

#### Backend: Go (Golang)

**Justificación:**
- Alto rendimiento y eficiencia (binario compilado)
- Excelente manejo de concurrencia (goroutines para checks paralelos)
- Bajo consumo de memoria y CPU
- Un solo binario ejecutable para desplegar
- Drivers nativos maduros para SQL Server y PostgreSQL
- Ideal para tareas de monitoreo y operaciones de red

**Framework:** net/http estándar (sin frameworks externos)

**Librerías:**
- `go-mssqldb`: Conexión a SQL Server
- `godotenv`: Carga de variables de entorno desde .env
- `google.golang.org/api/sheets/v4`: Google Sheets API v4
- `google.golang.org/api/option`: Opciones de autenticación para Google APIs
- `pgx/v5`: Driver PostgreSQL nativo con soporte completo para tipos y características de PostgreSQL

#### Frontend: React + Vite

**Justificación:**
- Build estático (sin necesidad de Node.js runtime en servidor)
- Vite ofrece desarrollo y builds extremadamente rápidos
- React permite UI/UX moderna y componentes reutilizables
- Renderizado en cliente (apropiado para dashboard interno)
- Menor consumo de recursos en servidor vs SSR

**Stack de UI:**
- React 18+
- Vite (build tool)
- TailwindCSS (estilos)
- shadcn/ui (componentes modernos basados en Radix UI)
- Recharts o similar (visualización de métricas)

**Alternativas descartadas:**
- Next.js: SSR innecesario para uso interno, requiere Node.js runtime adicional
- Solo Go con templates: Limitaciones para crear UI moderna e interactiva

### Arquitectura de Despliegue

```
Servidor Ubuntu (compartido):
  └─ Binario Go (proceso único)
      ├─ API REST (/api/*)
      │   ├─ Endpoints de estado de sistemas
      │   ├─ Endpoints de métricas de BD
      │   └─ WebSocket para updates en tiempo real
      │
      ├─ Servidor de archivos estáticos (/*)
      │   └─ Archivos compilados de React (HTML, CSS, JS)
      │
      ├─ Jobs de monitoreo en background
      │   ├─ Health checks HTTP
      │   ├─ Queries a bases de datos
      │   └─ Verificaciones de Google Apps Script
      │
      └─ Conexiones a recursos
          ├─ SQL Server (SaltaCompra prod/preprod)
          ├─ PostgreSQL (App.SaltaCompra)
          └─ Google APIs
```

**Ventajas de esta arquitectura:**
- Un solo proceso consumiendo recursos
- Deploy simple: copiar binario + carpeta de assets estáticos
- Sin dependencias de runtime (no requiere Node.js, Python, etc.)
- Escalable para múltiples sistemas futuros

### Estrategia de Monitoreo

**Métodos de verificación** (sin APIs de estado disponibles):
1. **HTTP Health Checks**: Peticiones a endpoints de aplicaciones
2. **Database Queries**: Consultas directas a BD para verificar actividad
3. **Verificación de registros**: Queries específicas según lógica de negocio
4. **RDAP Checks**: Consultas a servicio RDAP de NIC Argentina para monitorear expiración de dominios
5. **Google Sheets Checks**: Verificación de actualizaciones en hojas de cálculo vía Google Sheets API v4
6. **VPN Connectivity Checks**: Verificación de conectividad TCP a recursos internos antes de checks de BD

### Arquitectura de Sistemas

**Decisión: Separación por Ambiente**

Los sistemas se modelan como entidades independientes por ambiente (prod/preprod), en lugar de agruparlos por aplicación. Esta decisión se fundamenta en:

1. **Claridad de estado**: Permite identificar rápidamente si producción o preproducción tienen problemas
2. **Alertas granulares**: Diferentes niveles de criticidad (prod crítico, preprod menos urgente)
3. **Visualización en dashboard**: Tarjetas/cards separadas por ambiente
4. **Infraestructura compartida**: Elementos que afectan múltiples sistemas (dominio) se agrupan en un sistema independiente "Infraestructura Compartida"

**Sistemas implementados:**
- SaltaCompra Producción
- SaltaCompra Preproducción
- App.SaltaCompra (con verificación de VPN)
- Infraestructura Compartida (dominio)
- Google Sheets - Kairos Actualizaciones

### Decisiones de Implementación: Check de Mails

**Problema identificado:** El campo `sent_status` en SQL Server Database Mail es numérico, y no se conocían con certeza todos los posibles estados.

**Solución implementada:**

1. **Conversión de estados en la query:**
```sql
CASE sent_status
    WHEN 0 THEN 'unsent'
    WHEN 1 THEN 'sent'
    WHEN 3 THEN 'retrying'
    ELSE 'failed'
END as sent_status
```

2. **Conteo dinámico con mapa:** En lugar de hardcodear estados específicos, se usa `map[string]int` para contar automáticamente cualquier estado que aparezca en la BD.

**Ventajas:**
- Descubre automáticamente nuevos estados sin modificar código
- No requiere conocimiento previo de todos los valores posibles
- Escalable si SQL Server agrega estados en futuras versiones
- Metadata completa con `status_counts` para análisis detallado

**Tabla correcta:** `msdb.dbo.sysmail_mailitems` (no `sysmail_allitems` que es una vista diferente)

---

**ACTUALIZACIÓN: Análisis Diario Completo**

El check original solo analizaba los últimos 10 correos, lo cual era limitado para tener visibilidad real del servicio.

**Cambios implementados:**

1. **Análisis de todos los correos del día:**
   - Query filtra por `send_request_date = HOY` (con fallback a `last_mod_date`)
   - Procesa **todos** los registros del día

2. **Nuevas métricas:**
   - `today_total`, `today_sent`, `today_unsent`, `today_failed`, `today_retrying`
   - `today_failed_percentage`: Porcentaje de fallidos
   - `minutes_since_last_sent`: Tiempo desde último envío exitoso
   - `minutes_since_last_created`: Tiempo desde último registro agregado
   - `last_sent_time`, `last_created_time`: Timestamps precisos

3. **Sistema de alertas mejorado:**
   - **Error**: % fallidos >= umbral crítico (30%)
   - **Warning**: % fallidos >= umbral warning (10%)
   - **Warning**: Sin correos 'sent' en X horas (4 horas)
   - **OK**: Funcionamiento normal

**Configuración:**
```bash
MAIL_MAX_MINUTES_WITHOUT_SENT=240        # 4 horas
MAIL_DAILY_WARNING_FAILED_PERCENT=10     # 10%
MAIL_DAILY_ERROR_FAILED_PERCENT=30       # 30%
```

**Ventajas:**
- Visibilidad completa de actividad diaria
- Alertas proporcionales al volumen de correos
- Métricas temporales precisas
- Escalable a cualquier volumen

### Decisiones de Implementación: Check de Google Sheets (Kairos)

**Requisito:** Monitorear que la hoja "Log Actualizaciones Kairos" tenga un registro del día actual con la base de datos correcta.

**Solución implementada:**

1. **Autenticación flexible vía variables de entorno:**
   - Soporta Service Account (recomendado para backend)
   - Soporta API Key (para sheets públicos)
   - Ruta configurable al archivo de credenciales JSON

2. **Validaciones implementadas:**
   - Formato de fecha en TimeStamp (YYYY-MM-DD)
   - Formato de nombre de archivo (bd_YYYYMMDD.zip)
   - Coherencia entre fecha del TimeStamp y fecha extraída del nombre de archivo
   - Antigüedad de los datos (días desde última actualización)

3. **Estados del check:**
   - **ok**: Datos de hoy con BD correcta
   - **warning**: Datos de 1 día atrás (configurable vía `GSHEETS_WARNING_DAYS`)
   - **error**: Datos de 2+ días atrás o inconsistencias entre TimeStamp y archivo (configurable vía `GSHEETS_ERROR_DAYS`)

4. **Metadata proporcionada:**
   - `last_timestamp`: Fecha del último registro
   - `last_filename`: Nombre del archivo de BD utilizado
   - `filename_date`: Fecha extraída del nombre de archivo
   - `expected_filename`: Nombre de archivo esperado según TimeStamp
   - `days_old`: Días desde la última actualización
   - `total_rows`: Total de registros en la hoja (sin headers)

**Ventajas:**
- Configuración 100% vía variables de entorno (columnas, umbrales, credenciales)
- Validación doble para detectar inconsistencias
- Extensible para monitorear otras hojas de Google Sheets
- No requiere modificar código para cambiar parámetros de validación

### Decisiones de Implementación: Check de PostgreSQL con Verificación VPN

**Requisito:** Monitorear App.SaltaCompra (Next.js + PostgreSQL) que se encuentra dentro de una VPN privada.

**Problema identificado:** La base de datos PostgreSQL solo es accesible desde dentro de la red privada (VPN). Los checks fallarían si la VPN no está activa, dando errores confusos de timeout.

**Solución implementada:**

1. **Verificación previa de VPN:**
   - Función `CheckVPNConnectivity()`: Intenta conexión TCP al host con timeout corto (2 segundos por defecto)
   - Si la VPN no está disponible, el check falla inmediatamente con mensaje claro
   - Evita timeouts largos esperando conexiones que nunca se establecerán

2. **Check de PostgreSQL:**
   - Query de validación: `SELECT COUNT(*) FROM users.usuarios`
   - Verifica que la BD no solo esté online, sino que tenga datos reales
   - Metadata incluye: versión de PostgreSQL, conteo de usuarios, estado de VPN

3. **Configuración:**
   - Driver: `pgx/v5` (driver nativo de PostgreSQL para Go)
   - Host: `172.17.1.241` (IP interna en VPN)
   - Base de datos: `app_spc`
   - Timeout VPN: 2000ms (configurable vía `VPN_CHECK_TIMEOUT_MS`)

**Ventajas:**
- **Detección temprana**: Identifica problemas de VPN antes de intentar conectar a PostgreSQL
- **Mensajes claros**: El usuario sabe inmediatamente si el problema es VPN o BD
- **Sin timeouts largos**: Checks rápidos incluso cuando la VPN está caída
- **Metadata rica**: Incluye conteo de usuarios y versión de servidor
- **Extensible**: El mismo patrón puede usarse para otros recursos detrás de VPN

**Estructura del check:**
```go
{
  "status": "ok",
  "message": "Conexión PostgreSQL exitosa. 15 usuarios registrados (63ms)",
  "metadata": {
    "database": "app_spc",
    "server_version": "16.2",
    "user_count": 15,
    "vpn_available": true,
    "vpn_check_host": "172.17.1.241"
  }
}
```

### Decisiones de Implementación: Configuración Basada en Variables de Entorno

**Problema identificado:** URLs, dominios y contenidos esperados estaban hardcodeados en el código, dificultando cambios y testing.

**Solución implementada:**

1. **Nuevas estructuras de configuración en `config.go`:**
   - `SaltaCompraConfig`: URLs y contenido esperado para prod/preprod
   - `InfrastructureConfig`: Dominio y URL base de servicio RDAP

2. **Variables de entorno agregadas:**
   ```bash
   # SaltaCompra
   SALTACOMPRA_PROD_URL=https://saltacompra.gob.ar/
   SALTACOMPRA_PROD_EXPECTED_CONTENT=SALTA COMPRA - Portal de Compras...
   SALTACOMPRA_PREPROD_URL=https://preproduccion.saltacompra.gob.ar/
   SALTACOMPRA_PREPROD_EXPECTED_CONTENT=SALTA COMPRA - Portal de Compras...

   # Infraestructura
   INFRASTRUCTURE_DOMAIN=saltacompra.gob.ar
   RDAP_BASE_URL=https://rdap.nic.ar/domain/
   ```

3. **Refactoring realizado:**
   - `handlers.go`: Reemplazados valores hardcodeados por referencias a configuración
   - `rdap_domain_check.go`: URL base RDAP ahora configurable (mantiene fallback)
   - Todos los checks ahora reciben configuración vía structs

**Ventajas:**
- **Flexibilidad**: Cambiar URLs, dominios o contenido sin modificar código
- **Mantenibilidad**: Configuración centralizada en `.env`
- **Testing**: Fácil crear diferentes configuraciones para dev/staging/prod
- **Extensibilidad**: Agregar más dominios o servicios RDAP sin cambios de código
- **Reutilización**: Mismo código puede monitorear diferentes instancias

### Decisiones de Implementación: Manejo de Certificados SSL Autofirmados

**Problema identificado:** App.SaltaCompra usa certificado SSL autofirmado/CA privada, causando errores de verificación: `tls: failed to verify certificate: x509: certificate signed by unknown authority`

**Solución implementada:**

1. **Parámetro configurable `SkipSSLVerification`:**
   - Nuevo campo en `HTTPCheckConfig` para controlar verificación SSL por sistema
   - Configuración vía variable de entorno `APPSALTACOMPRA_SKIP_SSL_VERIFICATION`
   - Cuando está activo (`true`), el cliente HTTP usa `InsecureSkipVerify: true`

2. **Lógica de validación adaptativa:**
   - Si `SkipSSLVerification` está activo:
     - Cliente HTTP acepta cualquier certificado
     - No se ejecuta `validateSSLCertificate()` (que haría su propia petición HTTP con validación estricta)
     - Metadata incluye `"ssl_status": "skipped"` y `"ssl_verification_skipped": true`
   - Si está desactivado (comportamiento normal):
     - Cliente HTTP valida certificados contra CAs públicas
     - Se ejecuta validación SSL completa con cálculo de días hasta expiración

3. **Implementación técnica:**
   - `http_helpers.go`: `getHTTPClient(timeout, skipSSLVerify)` configura `TLSClientConfig` dinámicamente
   - `http_check.go`: Condicional `if ValidateSSL && !SkipSSLVerification` para ejecutar validación SSL
   - `config.go`: Nueva función `mustGetEnvAsBool()` para parsear booleanos desde `.env`
   - `handlers.go`: Pasa `SkipSSLVerification` desde configuración al check HTTP

**Ventajas:**
- **Selectivo por sistema**: Solo App.SaltaCompra saltea verificación, otros sistemas mantienen validación estricta
- **Seguridad**: Falla si variable no existe en `.env` (validación estricta por defecto)
- **Metadata transparente**: Indica explícitamente cuando verificación SSL fue salteada
- **Extensible**: Mismo patrón puede aplicarse a otros sistemas con certificados autofirmados

**Configuración:**
```bash
# .env
APPSALTACOMPRA_SKIP_SSL_VERIFICATION=true
```

## Estado del Proyecto

**Fase Actual**: MVP Backend funcional

### Implementado
- ✅ Estructura del monorepo (backend Go)
- ✅ Modelos de datos (System, Check)
- ✅ HTTP check para SaltaCompra (prod/preprod)
  - Validación de código HTTP
  - Verificación de contenido esperado
  - Validación de certificado SSL con días restantes
  - Umbrales de tiempo de respuesta (warning/error)
- ✅ DB check para servicio de correos (msdb.dbo.sysmail_mailitems)
  - Conteo dinámico de estados (sent, unsent, failed, retrying)
  - Conversión de sent_status numérico a string vía CASE
  - Detección de cola atascada (muchos unsent)
  - Metadata con breakdown completo por estado
- ✅ RDAP check para expiración de dominio
  - Consulta a API RDAP de NIC Argentina
  - Cálculo de días restantes hasta expiración
  - Alertas por umbrales configurables
- ✅ Google Sheets check para Kairos
  - Verificación de actualización diaria en "Log Actualizaciones Kairos"
  - Validación de coherencia entre TimeStamp y nombre de archivo de BD
  - Autenticación flexible (Service Account o API Key)
  - Umbrales configurables para warnings y errores
- ✅ PostgreSQL check para App.SaltaCompra
  - Verificación de VPN previa al check de BD
  - Query de validación con conteo de usuarios
  - Metadata con versión de servidor y estado de VPN
  - Detección temprana de problemas de conectividad
- ✅ Sistema "Infraestructura Compartida"
- ✅ Sistema "Google Sheets - Kairos Actualizaciones"
- ✅ Sistema "App.SaltaCompra"
- ✅ API REST: GET /api/systems
- ✅ Configuración 100% vía variables de entorno (.env)
  - URLs y contenidos esperados configurables
  - Dominios y servicios RDAP configurables
  - Credenciales PostgreSQL configurables
  - VPN check host y timeout configurables
  - Umbrales y parámetros de checks configurables
  - Archivo `.env.example` como template
  - Validación estricta sin fallbacks (fail-fast)
- ✅ Manejo de certificados SSL autofirmados/CA privadas
  - Configuración selectiva por sistema vía `SkipSSLVerification`
  - App.SaltaCompra configurado para aceptar certificados autofirmados
  - Metadata transparente cuando verificación SSL es salteada
- ✅ Servidor HTTP funcional

### Configuración

**Archivo `.env` (base de toda la configuración):**

1. **URLs y Contenidos:**
   - `SALTACOMPRA_PROD_URL`: URL de producción
   - `SALTACOMPRA_PROD_EXPECTED_CONTENT`: Texto esperado en HTML de prod
   - `SALTACOMPRA_PREPROD_URL`: URL de preproducción
   - `SALTACOMPRA_PREPROD_EXPECTED_CONTENT`: Texto esperado en HTML de preprod

2. **Infraestructura:**
   - `INFRASTRUCTURE_DOMAIN`: Dominio a monitorear
   - `RDAP_BASE_URL`: URL base del servicio RDAP (ej: https://rdap.nic.ar/domain/)

3. **Credenciales:**
   - SQL Server prod/preprod: host, port, user, password, database
   - PostgreSQL App.SaltaCompra: host, port, user, password, database
   - Google Sheets: spreadsheet ID, sheet name, auth method, credentials file path

4. **VPN:**
   - `VPN_CHECK_HOST`: Host para verificar conectividad VPN
   - `VPN_CHECK_TIMEOUT_MS`: Timeout en ms para verificación VPN (default: 2000)

5. **Umbrales de monitoreo:**
   - Mails: 180 min sin envío = warning, >5 fallidos = error, >7 unsent = warning (cola atascada)
   - SSL: 30 días antes = warning
   - HTTP: 3000ms = warning, 10000ms = error
   - Dominio: 60 días = warning, 30 días = error
   - Google Sheets: 1 día = warning, 2 días = error

**Archivos complementarios:**
- `.env.example`: Template con todas las variables disponibles
- `./credentials/service-account.json`: Service Account de Google (gitignored)
- Usuario SQL: `readOnlyUser` con permisos SELECT en msdb.dbo.sysmail_mailitems

### Decisiones de Implementación: Arquitectura de Cache + SSE + Background Worker

**Requisito:** Minimizar impacto en bases de datos de producción y proporcionar updates en tiempo real al frontend sin esperas largas.

**Problema identificado:**
- Ejecutar checks síncronos en cada petición HTTP causaría timeouts y carga excesiva en BD
- No se puede esperar 5-10 segundos por respuesta en cada request
- Necesidad de updates progresivos en UI (no esperar a que todos los checks terminen)

**Solución implementada:**

#### 1. **Sistema de Cache Thread-Safe** (`internal/cache/system_cache.go`)

Cache en memoria con sincronización para almacenar estado de sistemas:

```go
type SystemCache struct {
    mu      sync.RWMutex
    systems map[string]CachedSystem
}

type CachedSystem struct {
    Data      models.System
    UpdatedAt time.Time
}
```

**Métodos:**
- `Get(id)`: Obtiene un sistema del cache
- `Set(id, system)`: Actualiza un sistema en el cache
- `GetAll()`: Retorna todos los sistemas cacheados
- `IsStale(id, maxAge)`: Verifica si un sistema está desactualizado
- Thread-safe con `sync.RWMutex` para lectura/escritura concurrente

#### 2. **Broadcaster SSE** (`internal/sse/broadcaster.go`)

Sistema de Server-Sent Events para enviar updates en tiempo real al frontend:

```go
type Broadcaster struct {
    mu      sync.RWMutex
    clients map[string]*Client
}
```

**Eventos SSE:**
- `connected`: Cliente conectado (con client_id)
- `system_update`: Actualización de un sistema específico
- `check_complete`: Todos los checks completaron

**Características:**
- Múltiples clientes simultáneos
- Broadcast selectivo por evento
- Buffer de 10 mensajes por cliente
- Auto-cleanup al desconectar

#### 3. **Smart Background Worker** (`internal/scheduler/smart_worker.go`)

Worker inteligente que ejecuta checks periódicamente con lógica de pausa:

```go
type SmartWorker struct {
    interval        time.Duration  // Intervalo entre checks
    idleTimeout     time.Duration  // Timeout para pausar si no hay actividad
    lastActivity    time.Time      // Última request recibida
}
```

**Comportamiento:**
- Ejecuta checks cada **30 minutos** (configurable)
- Se **pausa automáticamente** si no hay actividad por **60 minutos**
- Se **reactiva** cuando llega una nueva request
- Actualiza cache y envía eventos SSE al completar checks

**Ventajas:**
- Minimiza carga en BD de producción (solo checks cada 30 min)
- Ahorra recursos cuando nadie usa el dashboard
- Datos siempre frescos si hay usuarios activos

#### 4. **Endpoints API Refactorizados**

**GET `/api/systems`** - Lectura instantánea del cache:
```go
// Responde en <10ms, siempre disponible
// Retorna: {"systems": [...], "cached": true, "count": 5}
```

**GET `/api/events`** - Stream SSE para updates en tiempo real:
```go
// Conexión persistente HTTP
// Recibe eventos conforme se actualizan sistemas
// event: system_update → Datos de un sistema
// event: check_complete → Todos los checks terminaron
```

**POST `/api/refresh`** - Refresh manual de todos los sistemas:
```go
// Responde 202 Accepted inmediatamente
// Ejecuta checks en background
// Envía updates vía SSE conforme completan
```

**POST `/api/systems/:id/refresh`** - Refresh de sistema individual:
```go
// Responde 202 Accepted inmediatamente
// Ejecuta check del sistema específico
// Envía update vía SSE al completar
```

#### 5. **Broadcasting Progresivo**

Cada check envía evento SSE **inmediatamente al completar** (no espera a los demás):

```go
func checkAllSystemsWithProgressiveBroadcast() {
    // Ejecuta 5 checks en paralelo
    go func() {
        system := checkSaltaCompraProd()
        cache.Set(system.ID, system)
        broadcaster.BroadcastSystem(system)  // ← Envío inmediato
    }()
    // ... otros 4 sistemas en paralelo
}
```

**Resultado:** Frontend ve sistemas aparecer progresivamente (1-2 segundos de diferencia), no todos juntos.

#### 6. **Middleware CORS**

Headers CORS configurados para permitir consumo desde frontend:
```go
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type
```

#### 7. **Configuración**

Variables de entorno agregadas:
```bash
# Background Worker
BACKGROUND_CHECK_INTERVAL_MINUTES=30  # Checks automáticos cada 30 min
WORKER_IDLE_TIMEOUT_MINUTES=60        # Pausar si idle por 60 min
CACHE_MAX_AGE_MINUTES=35              # Validez de datos en cache
```

**Ventajas de esta arquitectura:**
- ✅ **Respuesta instantánea**: API responde en <10ms (cache)
- ✅ **Updates en tiempo real**: SSE envía datos conforme se obtienen
- ✅ **Mínimo impacto en BD**: Solo checks cada 30 min (si hay actividad)
- ✅ **Escalable**: Soporta múltiples clientes SSE simultáneos
- ✅ **Eficiente**: Worker se pausa cuando no se usa
- ✅ **Progresivo**: UI ve sistemas aparecer uno por uno
- ✅ **Asíncrono**: Refresh manual no bloquea (202 Accepted)

**Flujo completo:**
```
1. Usuario abre dashboard
   → GET /api/systems (responde cache instantáneo)
   → EventSource('/api/events') (abre SSE)

2. Si cache vacío/viejo
   → POST /api/refresh (dispara checks)
   → Backend ejecuta checks en paralelo
   → Cada check envía SSE al completar
   → Frontend actualiza UI progresivamente

3. Background worker (cada 30 min)
   → Si hay actividad reciente, ejecuta checks
   → Actualiza cache
   → Envía eventos SSE a clientes conectados
   → Si idle >60 min, se pausa
```

### Frontend Implementado

**Fase Actual**: MVP Completo - Backend + Frontend 100% funcional

#### Stack Tecnológico Frontend
- ✅ React 18 + TypeScript
- ✅ Vite 7 (build tool ultra-rápido)
- ✅ TailwindCSS v4 (estilos utility-first)
- ✅ Radix UI (componentes accesibles headless)
- ✅ Lucide React (iconos)
- ✅ EventSource API nativa (SSE)
- ✅ Fetch API nativa (HTTP)

#### Componentes Implementados (`frontend/src/components/`)

**1. `Dashboard.tsx`** - Layout principal (~90 líneas)
- Integra todos los componentes
- Maneja estados de loading y error
- Skeleton de carga
- Empty states

**2. `Header.tsx`** - Cabecera (~60 líneas)
- Logo + título
- Indicador de conexión SSE (punto verde pulsante)
- Timestamp de última actualización
- Botón "Refrescar" con spinner

**3. `StatsOverview.tsx`** - Resumen de métricas (~50 líneas)
- 4 cards de estadísticas: Total, Online, Warnings, Errores
- Cálculo de porcentajes
- Colores dinámicos por estado
- Helper component: `StatCard` (co-located)

**4. `SystemCard.tsx`** - Cards de sistemas expandibles (~180 líneas)
- Accordion con Radix UI (expandir/colapsar)
- Header siempre visible con:
  - Icono de estado (emoji)
  - Nombre del sistema
  - Badge de ambiente (prod/preprod/shared)
  - Badge de origen de datos (Cache/Actualizando/En vivo)
  - Tiempo de último check
  - Tiempo de respuesta promedio
  - Contador de checks
- Contenido expandible con:
  - Lista de checks con metadata detallada
  - Colores por estado
  - Response time por check
- Animaciones al actualizar por SSE:
  - Flash (opacity pulse 0.5s)
  - Pulse ring (sombra verde pulsante 1s)
  - Fade highlight (background que desaparece 2s)
- Helper component: `CheckItem` (co-located)

**5. `DataSourceBadge.tsx`** - Indicador origen de datos (~65 líneas)
- **"Cache"** (gris, icono Database) - Datos del cache
- **"Actualizando..."** (amarillo, icono RefreshCw spinning) - Refresh en curso
- **"En vivo"** (verde pulsante, icono Radio) - Actualizado por SSE
- Lógica de prioridad para mostrar estado correcto

**6. `ProgressBar.tsx`** - Barra de progreso de actualización (~60 líneas)
- Muestra "Actualizando sistemas: X/5"
- Barra visual con porcentaje animado
- Icono RefreshCw que gira durante actualización
- Cambia a verde al completar 100%
- Se oculta automáticamente después de 500ms

#### Hooks Personalizados (`frontend/src/hooks/`)

**1. `useSSE.ts`** - Manejo de conexión SSE (~70 líneas)
- Envuelve EventSource API nativa
- Maneja eventos: `connected`, `system_update`, `check_complete`
- Auto-cleanup al desmontar
- Estado de conexión (`connected`, `clientId`)
- Métodos: `connect()`, `disconnect()`, `isConnected()`

**2. `useSystems.ts`** - Estado principal de sistemas (~140 líneas)
- **Estrategia híbrida**:
  1. Carga instantánea del cache (0ms)
  2. Conecta SSE automáticamente
  3. Auto-refresh al conectar SSE (sin intervención del usuario)
- **Rastreo de origen**: Marca cada sistema como `source: 'cache' | 'sse'`
- **Progreso de actualización**: `refreshProgress: { updated: number, total: number }`
- **Estados**: `loading`, `error`, `cached`, `refreshing`, `sseConnected`
- **Actualización incremental**: Contador se incrementa con cada `system_update` recibido
- **Reseteo automático**: Al completar todos los sistemas, resetea `refreshing` después de 500ms

#### Servicios (`frontend/src/services/`)

**1. `api.ts`** - Cliente HTTP (~50 líneas)
- `getSystems()` - GET /api/systems (cache)
- `refreshAllSystems()` - POST /api/refresh
- `refreshSystem(id)` - POST /api/systems/:id/refresh
- Usa Fetch API nativa (sin librerías externas)

**2. `sse.ts`** - Cliente SSE (~100 líneas)
- Clase `SSEClient` que envuelve EventSource
- Callbacks tipados para cada evento
- Manejo de errores
- Cleanup automático
- Log de eventos en consola

#### Utilidades (`frontend/src/lib/utils.ts`) (~200 líneas)

- `cn()` - Merge de clases Tailwind con clsx + tailwind-merge
- `getStatusClasses()` - Colores por estado (online/warning/error)
- `getStatusIcon()` - Emojis por estado
- `getEnvironmentBadge()` - Estilos por ambiente
- `formatRelativeTime()` - "hace 2 min" en español
- `formatResponseTime()` - "127ms"
- `getAverageResponseTime()` - Calcula promedio de checks
- `calculateSystemStats()` - Estadísticas agregadas

#### Tipos (`frontend/src/types/system.ts`) (~65 líneas)

Interfaces TypeScript que coinciden exactamente con modelos Go:
- `System` (con campos extras: `source`, `localUpdatedAt`)
- `Check`
- `SystemsResponse`
- `RefreshResponse`
- `SSEConnectedEvent`
- `SSECheckCompleteEvent`
- `SystemStats`

#### Configuración

**`tailwind.config.js`** - Tema personalizado:
- Paleta de colores: `success`, `warning`, `error` (50-900)
- Animaciones:
  - `fadeIn` - Fade in con translateY
  - `slideDown` / `slideUp` - Para Accordion
  - `pulse` - Pulso estándar
  - `flash` - Flash de opacity (0.5s)
  - `pulse-ring` - Pulso de sombra (1s)
  - `fade-highlight` - Fade de background (2s)
  - `pulse-subtle` - Pulso sutil infinito (2s)

**`postcss.config.js`** - Procesador CSS:
- `@tailwindcss/postcss` (requerido para Tailwind v4)
- `autoprefixer`

**`index.css`** - Estilos base:
- `@import "tailwindcss"` (sintaxis Tailwind v4)
- Estilos de body con fuentes del sistema

#### Flujo de Usuario Completo

```
1. Usuario abre http://localhost:5173
   ├─ GET /api/systems (cache instantáneo)
   ├─ Muestra loading skeleton
   ├─ Renderiza dashboard con datos del cache
   └─ Badges: "Cache" (gris)

2. SSE se conecta automáticamente (~500ms)
   ├─ EventSource('/api/events')
   ├─ Recibe evento: connected
   └─ Hook detecta conexión

3. Auto-refresh se dispara (automático, sin click)
   ├─ POST /api/refresh (202 Accepted)
   ├─ Barra aparece: "Actualizando sistemas: 0/5" [0%]
   ├─ Badges cambian a: "Actualizando..." (amarillo, spinner)
   └─ Checks se ejecutan en paralelo en backend

4. Sistemas se actualizan progresivamente (~1-10 seg)
   ├─ SSE envía: system_update (Sistema 1)
   │   ├─ Card hace: flash + pulse ring + fade highlight
   │   ├─ Badge cambia a: "En vivo" (verde pulsante)
   │   └─ Barra: "1/5" [20%]
   ├─ SSE envía: system_update (Sistema 2)
   │   ├─ Animaciones visuales
   │   ├─ Badge: "En vivo"
   │   └─ Barra: "2/5" [40%]
   └─ ... (continúa hasta 5/5)

5. Actualización completa (~5-10 seg total)
   ├─ Barra: "5/5" [100%] (verde)
   ├─ SSE envía: check_complete
   ├─ Después de 500ms: barra desaparece con fade
   ├─ Todos los badges: "En vivo" (pulsantes)
   └─ Botón "Refrescar" listo para uso manual
```

#### Características Frontend

✅ **Carga instantánea** - Cache en <10ms
✅ **Auto-refresh automático** - Al conectar SSE
✅ **Updates en tiempo real** - Vía SSE sin polling
✅ **Indicadores de origen** - Cache/Actualizando/En vivo
✅ **Barra de progreso** - "X/5 sistemas" con porcentaje
✅ **Animaciones visuales** - Flash, pulse ring, fade highlight
✅ **Cards expandibles** - Accordion sin routing
✅ **Metadata detallada** - Muestra hasta 6 campos por check
✅ **Responsive design** - Grid adaptativo
✅ **Loading states** - Skeleton, error states, empty states
✅ **Colores dinámicos** - Verde/amarillo/rojo por estado
✅ **Timestamps relativos** - "hace X min" en español
✅ **Botón manual** - Refresh forzado disponible
✅ **Zero runtime overhead** - Sin Redux, Context, styled-components
✅ **Minimal bundle** - ~570 líneas, <150KB gzipped

#### Gestión de Paquetes

**pnpm** configurado como manejador de paquetes:
- `pnpm-workspace.yaml` - Configuración de workspace
- `package.json` (raíz) - Scripts de ejecución:
  - `pnpm backend` - Ejecuta solo backend
  - `pnpm frontend` - Ejecuta solo frontend
  - `pnpm dev` - Ejecuta ambos con concurrently
  - `pnpm build` - Build del frontend
  - `pnpm install:all` - Instala deps de raíz + frontend
- `.gitignore` actualizado para pnpm

### Pendiente
- Autenticación y seguridad
- Historial de checks/logs
- Notificaciones/alertas
- Configuración de umbrales desde UI
- Deploy a producción

## Desarrollo y Testing

### Ejecución en Desarrollo

**RECOMENDADO**: Usar los scripts desde la raíz del proyecto con pnpm:

```bash
# Ejecutar backend + frontend simultáneamente
pnpm dev

# O ejecutarlos por separado:
pnpm backend   # Solo backend en localhost:8080
pnpm frontend  # Solo frontend en localhost:5173
```

**Ejecución directa (alternativa):**

```bash
# Backend
cd backend
go run ./cmd/server/main.go

# Frontend
cd frontend
pnpm dev
```

**Razones para usar `go run`:**
- Cambios se reflejan inmediatamente sin necesidad de compilar
- Más rápido para iterar durante desarrollo
- Evita tener binarios desactualizados
- Facilita el debugging

**Nota**: La compilación (`go build`) solo debe usarse para preparar el binario de producción o cuando se requiera específicamente.

### Validación de Configuración

El servidor implementa validación estricta de configuración:
- **Archivo `.env` obligatorio**: Si no existe, el servidor falla inmediatamente
- **Variables requeridas**: Todas las variables de configuración son obligatorias
- **Sin fallbacks**: No hay valores por defecto, se debe configurar explícitamente cada variable
- **Fail-fast**: El servidor no arranca si falta alguna configuración crítica

Si el servidor no arranca, revisar los mensajes de error que indicarán exactamente qué variable falta.

## Notas Técnicas

Este documento será actualizado continuamente conforme se tomen decisiones técnicas fundamentadas y se descubran nuevos requisitos durante el proceso de desarrollo.

**Ver también**: `TODO.md` para lista de tareas pendientes y mejoras futuras.
