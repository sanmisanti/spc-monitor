# Sistema de Monitoreo de Infraestructura

Sistema de monitoreo centralizado para la Secretaría de Contrataciones. Permite monitorear el estado operacional de todos los sistemas bajo responsabilidad del área de ingeniería informática.

## 🚀 Inicio Rápido

### Requisitos Previos

- **Go** 1.21+ (para el backend)
- **Node.js** 18+ (para el frontend)
- **pnpm** 9+ (gestor de paquetes)
- **Git**

**Instalar pnpm (si no lo tienes):**
```bash
npm install -g pnpm
```

### Instalación

1. **Clonar el repositorio:**
   ```bash
   git clone <repository-url>
   cd monitor
   ```

2. **Instalar todas las dependencias:**
   ```bash
   pnpm install:all
   # Esto instala dependencias de raíz y del frontend
   ```

3. **Configurar variables de entorno:**
   ```bash
   # Copiar y editar el archivo .env del backend
   cp backend/.env.example backend/.env
   # Editar backend/.env con tus credenciales
   ```

## 📦 Comandos Disponibles

Desde la **raíz del proyecto**, puedes ejecutar:

### 🔹 Backend (Go)
```bash
pnpm backend
```
Ejecuta el servidor Go en `http://localhost:8080`

### 🔹 Frontend (React + Vite)
```bash
pnpm frontend
```
Ejecuta el servidor de desarrollo Vite en `http://localhost:5173`

### 🔹 Ambos (Backend + Frontend) ⭐
```bash
pnpm dev
```
Ejecuta backend y frontend simultáneamente con salida coloreada.

**Nota:** Para detener ambos procesos, presiona `Ctrl+C` una vez (detendrá ambos automáticamente).

### 🔹 Build Frontend para Producción
```bash
pnpm build
```
Compila el frontend para producción en `frontend/dist/`

### 🔹 Instalar Dependencias
```bash
pnpm install:all        # Instala deps de raíz + frontend
pnpm install:frontend   # Solo frontend
```

---

## 📂 Estructura del Proyecto

```
monitor/
├── backend/                # Servidor Go
│   ├── cmd/server/        # Punto de entrada
│   ├── internal/          # Código interno
│   │   ├── api/          # Handlers HTTP
│   │   ├── cache/        # Sistema de cache
│   │   ├── config/       # Configuración
│   │   ├── models/       # Modelos de datos
│   │   ├── monitors/     # Checks de sistemas
│   │   ├── scheduler/    # Background worker
│   │   └── sse/          # Server-Sent Events
│   ├── .env              # Variables de entorno (no commitear)
│   └── go.mod            # Dependencias Go
│
├── frontend/              # Aplicación React
│   ├── src/
│   │   ├── components/   # Componentes React
│   │   ├── hooks/        # Custom hooks
│   │   ├── services/     # API + SSE clients
│   │   ├── types/        # TypeScript types
│   │   └── lib/          # Utilidades
│   └── package.json      # Dependencias npm
│
├── package.json          # Scripts del proyecto
├── CLAUDE.md             # Documentación técnica
├── TODO.md               # Tareas pendientes
└── README.md             # Este archivo
```

---

## 🔧 Desarrollo

### Backend (Go)

**Ejecución directa:**
```bash
cd backend
go run ./cmd/server/main.go
```

**Compilar binario:**
```bash
cd backend
go build -o monitor.exe ./cmd/server
./monitor.exe
```

**Variables de entorno importantes:**
- Ver `backend/.env.example` para la lista completa
- Configurar credenciales de BD, URLs de sistemas, etc.

### Frontend (React + TypeScript)

**Ejecución directa:**
```bash
cd frontend
pnpm dev
```

**Build para producción:**
```bash
cd frontend
pnpm build
# Los archivos compilados estarán en frontend/dist/
```

**Tecnologías:**
- React 18 + TypeScript
- Vite (build tool)
- TailwindCSS (estilos)
- Radix UI (componentes)
- Lucide React (iconos)

---

## 🌐 URLs

| Servicio | URL | Descripción |
|----------|-----|-------------|
| **Frontend** | http://localhost:5173 | Dashboard de monitoreo |
| **Backend API** | http://localhost:8080 | API REST |
| **SSE Stream** | http://localhost:8080/api/events | Server-Sent Events |

### Endpoints de la API

- `GET /api/systems` - Lista de sistemas (cache)
- `GET /api/events` - Stream SSE de updates en tiempo real
- `POST /api/refresh` - Refresh manual de todos los sistemas
- `POST /api/systems/:id/refresh` - Refresh de sistema individual
- `GET /api/health` - Health check

---

## 🎯 Características

### Backend
✅ Monitoreo de múltiples tipos de sistemas (HTTP, BD, RDAP, Google Sheets)
✅ Cache thread-safe en memoria
✅ Server-Sent Events para updates en tiempo real
✅ Background worker inteligente (se pausa si no hay actividad)
✅ Checks en paralelo con broadcasts progresivos
✅ Configuración 100% vía variables de entorno

### Frontend
✅ Dashboard moderno con React + TypeScript
✅ Carga instantánea desde cache
✅ Auto-refresh automático al conectar
✅ Updates en tiempo real vía SSE
✅ Indicadores de origen de datos (Cache/Actualizando/En vivo)
✅ Barra de progreso de actualización
✅ Animaciones visuales en cada update
✅ Cards expandibles con metadata detallada
✅ Responsive design con TailwindCSS

---

## 📊 Sistemas Monitoreados

1. **SaltaCompra Producción** (ASP.NET + SQL Server)
2. **SaltaCompra Preproducción** (ASP.NET + SQL Server)
3. **App.SaltaCompra** (Next.js + PostgreSQL)
4. **Google Sheets - Kairos** (Apps Script)
5. **Infraestructura Compartida** (Dominio)

Ver `CLAUDE.md` para detalles técnicos completos.

---

## 🐛 Troubleshooting

### Backend no arranca
- Verificar que el archivo `backend/.env` existe y está configurado
- Verificar que el puerto 8080 no está en uso
- Revisar credenciales de BD en `.env`

### Frontend no arranca
- Ejecutar `pnpm install:frontend` para instalar dependencias
- Verificar que el puerto 5173 no está en uso
- Limpiar cache: `cd frontend && rm -rf node_modules pnpm-lock.yaml && pnpm install`

### Backend y Frontend no se comunican
- Verificar que el backend esté corriendo en `http://localhost:8080`
- Revisar configuración de CORS en `backend/cmd/server/main.go`
- Verificar URL de API en `frontend/src/services/api.ts`

---

## 📝 Documentación

- **`CLAUDE.md`** - Documentación técnica completa, decisiones de arquitectura
- **`TODO.md`** - Lista de tareas pendientes y mejoras futuras
- **`DEPLOY.md`** - Guía de despliegue a producción

---

## 🤝 Contribución

Este es un proyecto interno de la Secretaría de Contrataciones. Para contribuir:

1. Crear una rama desde `main`
2. Implementar cambios siguiendo las guías en `CLAUDE.md`
3. Probar localmente con `pnpm dev`
4. Crear Pull Request con descripción detallada

---

## 📄 Licencia

Uso interno - Secretaría de Contrataciones de la Provincia de Salta.
