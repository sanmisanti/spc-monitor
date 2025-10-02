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
  - Producción (servidor dedicado + BD propia)
  - Preproducción (servidor dedicado + BD propia)

### 2. App.SaltaCompra
- **Tecnología**: Next.js
- **Infraestructura**: Servidor propio Ubuntu
- **Base de Datos**: PostgreSQL

### 3. Sistema Google Apps Script
- **Tecnología**: Google Apps Script
- **Infraestructura**: Google Cloud Platform
- **Storage**: Google Spreadsheets

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
- `pgx`: Driver PostgreSQL (pendiente)
- `google-api-go-client`: Google APIs (pendiente)

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

## Estado del Proyecto

**Fase Actual**: MVP Backend funcional

### Implementado
- ✅ Estructura del monorepo (backend Go)
- ✅ Modelos de datos (System, Check)
- ✅ HTTP check para SaltaCompra (prod/preprod)
- ✅ DB check para servicio de correos (msdb.dbo.sysmail_allitems)
- ✅ API REST: GET /api/systems
- ✅ Configuración vía variables de entorno (.env)
- ✅ Servidor HTTP funcional

### Configuración
- Variables en `.env`: credenciales SQL Server, umbrales de monitoreo
- Usuario SQL: `readOnlyUser` con permisos SELECT en msdb.dbo.sysmail_allitems
- Umbrales: 180 min sin mails = warning, >5 mails fallidos = error

### Pendiente
- Frontend React + Vite
- Checks para App.SaltaCompra (PostgreSQL)
- Checks para Google Apps Script
- Jobs automáticos periódicos (scheduler)
- WebSocket para updates en tiempo real
- Autenticación y seguridad

## Notas Técnicas

Este documento será actualizado continuamente conforme se tomen decisiones técnicas fundamentadas y se descubran nuevos requisitos durante el proceso de desarrollo.
