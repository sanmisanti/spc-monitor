# Guía de Despliegue

## Scripts Disponibles

### 1. `deploy-simple.bat` (Recomendado para uso regular)
Script de un solo clic. Ejecuta el despliegue con configuración predeterminada.

**Uso:**
```cmd
deploy-simple.bat
```

### 2. `deploy.ps1` (PowerShell avanzado)
Script principal con opciones personalizables.

**Uso básico:**
```powershell
.\deploy.ps1
```

**Uso con parámetros personalizados:**
```powershell
.\deploy.ps1 -ServerIP "192.168.49.64" -ShareName "spc-monitor" -Username "user_deploy" -Password "deploy_monitor"
```

## Configuración del Servidor

### Servidor destino:
- **IP**: 192.168.49.64
- **Carpeta compartida**: `\\192.168.49.64\spc-monitor`
- **Usuario**: `user_deploy`
- **Contraseña**: `deploy_monitor`

### Ruta local en el servidor:
```
C:\Users\usuario\proyectos\spc-monitor
```

## Qué Hace el Script

1. **Conecta** a la carpeta compartida con autenticación
2. **Limpia** la carpeta destino (elimina archivos antiguos)
3. **Copia** todos los archivos del proyecto excepto:
   - Scripts de despliegue (`deploy.ps1`, `deploy-simple.bat`)
   - Archivos temporales
   - `.git`, `.vscode`
   - `node_modules` (si existe)
   - Binarios `.exe` (se recompilan en el servidor)
4. **Muestra** resumen de archivos copiados
5. **Desconecta** la unidad de red

## Archivos Copiados

El script copia:
- ✅ Todo el código fuente Go (`backend/`)
- ✅ Archivos de configuración (`.env`, `go.mod`, `go.sum`)
- ✅ Credenciales (`backend/credentials/*.json`)
- ✅ Documentación (`CLAUDE.md`, `TODO.md`)
- ✅ Estructura completa de directorios

## Seguridad

- La conexión requiere autenticación con usuario y contraseña
- Solo el usuario `user_deploy` tiene permisos de escritura en la carpeta compartida
- Las credenciales están incluidas en el script (considera moverlas a variables de entorno para mayor seguridad)

## Solución de Problemas

### Error: "Acceso denegado"
Verifica que:
- El usuario `user_deploy` existe en el servidor
- La carpeta está compartida correctamente: `net share spc-monitor`
- El firewall permite compartición de archivos

### Error: "No se encuentra el nombre de red"
Verifica que:
- El servidor está encendido y conectado a la red
- La IP del servidor es correcta: `ping 192.168.49.64`
- El servicio de red está activo en el servidor: `Get-Service LanmanServer`

### El script no se ejecuta
Si PowerShell bloquea la ejecución, ejecuta como administrador:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

## Próximos Pasos Después del Despliegue

Una vez copiados los archivos en el servidor:

1. **Compilar el proyecto Go:**
   ```bash
   cd C:\Users\usuario\proyectos\spc-monitor\backend
   go build -o monitor.exe ./cmd/server
   ```

2. **Configurar el archivo `.env`** con los valores específicos del servidor

3. **Ejecutar el servidor:**
   ```bash
   .\monitor.exe
   ```

4. **Configurar como servicio de Windows** (opcional pero recomendado para producción)
