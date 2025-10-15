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
- **Tecnología**: Next.js
- **Infraestructura**: VM Ubuntu dedicada
- **Base de Datos**: PostgreSQL
- **Ubicación**: Dentro de VPN (requiere conexión VPN para checks internos)
- **Nota**: Servidor contiene web + base de datos + otros sistemas a explorar

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
- `pgx`: Driver PostgreSQL (pendiente)

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
- **Valores por defecto**: Config tiene fallbacks para facilitar desarrollo

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
- ✅ Sistema "Infraestructura Compartida"
- ✅ Sistema "Google Sheets - Kairos Actualizaciones"
- ✅ API REST: GET /api/systems
- ✅ Configuración 100% vía variables de entorno (.env)
  - URLs y contenidos esperados configurables
  - Dominios y servicios RDAP configurables
  - Umbrales y parámetros de checks configurables
  - Archivo `.env.example` como template
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
   - Google Sheets: spreadsheet ID, sheet name, auth method, credentials file path

4. **Umbrales de monitoreo:**
   - Mails: 180 min sin envío = warning, >5 fallidos = error, >7 unsent = warning (cola atascada)
   - SSL: 30 días antes = warning
   - HTTP: 3000ms = warning, 10000ms = error
   - Dominio: 60 días = warning, 30 días = error
   - Google Sheets: 1 día = warning, 2 días = error

**Archivos complementarios:**
- `.env.example`: Template con todas las variables disponibles
- `./credentials/service-account.json`: Service Account de Google (gitignored)
- Usuario SQL: `readOnlyUser` con permisos SELECT en msdb.dbo.sysmail_mailitems

### Pendiente
- Frontend React + Vite
- Checks para App.SaltaCompra (PostgreSQL)
- Jobs automáticos periódicos (scheduler)
- WebSocket para updates en tiempo real
- Autenticación y seguridad

## Notas Técnicas

Este documento será actualizado continuamente conforme se tomen decisiones técnicas fundamentadas y se descubran nuevos requisitos durante el proceso de desarrollo.

**Ver también**: `TODO.md` para lista de tareas pendientes y mejoras futuras.
