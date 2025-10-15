# Script de Despliegue - Sistema de Monitoreo
# Copia todo el proyecto al servidor de producción

param(
    [string]$ServerIP = "192.168.49.64",
    [string]$ShareName = "spc-monitor",
    [string]$Username = "user_deploy",
    [string]$Password = "deploy_monitor"
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Script de Despliegue - Monitor SPC" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Ruta del proyecto local
$ProjectPath = $PSScriptRoot
Write-Host "Ruta del proyecto: $ProjectPath" -ForegroundColor Yellow

# Crear credenciales
$SecurePassword = ConvertTo-SecureString $Password -AsPlainText -Force
$Credential = New-Object System.Management.Automation.PSCredential($Username, $SecurePassword)

# Nombre de la unidad temporal
$DriveName = "DeployDrive"

try {
    # Conectar a la carpeta compartida
    Write-Host ""
    Write-Host "Conectando a \\$ServerIP\$ShareName..." -ForegroundColor Yellow

    # Remover unidad si ya existe
    if (Get-PSDrive -Name $DriveName -ErrorAction SilentlyContinue) {
        Remove-PSDrive -Name $DriveName -Force
    }

    New-PSDrive -Name $DriveName -PSProvider FileSystem -Root "\\$ServerIP\$ShareName" -Credential $Credential | Out-Null
    Write-Host "Conexión exitosa!" -ForegroundColor Green

    # Limpiar carpeta destino (opcional - comentar si no se desea limpiar)
    Write-Host ""
    Write-Host "Limpiando carpeta destino..." -ForegroundColor Yellow
    Get-ChildItem "${DriveName}:" -Recurse | Remove-Item -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host "Carpeta destino limpiada" -ForegroundColor Green

    # Copiar archivos
    Write-Host ""
    Write-Host "Copiando archivos..." -ForegroundColor Yellow

    # Archivos y carpetas a excluir
    $Exclude = @(
        "test_connection.ps1",  # Script de prueba temporal
        "deploy.ps1",           # Este mismo script
        ".git",                 # Repositorio git
        ".vscode",              # Configuración de VSCode
        "node_modules",         # Dependencias de Node (si existen)
        "*.exe"                 # Binarios compilados (se recompilan en el servidor)
    )

    # Copiar con progreso
    $FilesToCopy = Get-ChildItem -Path $ProjectPath -Recurse -File | Where-Object {
        $file = $_
        $shouldExclude = $false

        foreach ($pattern in $Exclude) {
            if ($file.Name -like $pattern -or $file.FullName -like "*\$pattern\*") {
                $shouldExclude = $true
                break
            }
        }

        -not $shouldExclude
    }

    $TotalFiles = $FilesToCopy.Count
    $Counter = 0

    foreach ($File in $FilesToCopy) {
        $Counter++
        $RelativePath = $File.FullName.Substring($ProjectPath.Length + 1)
        $DestPath = Join-Path "${DriveName}:" $RelativePath
        $DestDir = Split-Path $DestPath -Parent

        # Crear directorio si no existe
        if (-not (Test-Path $DestDir)) {
            New-Item -ItemType Directory -Path $DestDir -Force | Out-Null
        }

        # Copiar archivo
        Copy-Item -Path $File.FullName -Destination $DestPath -Force

        # Mostrar progreso
        $Percent = [math]::Round(($Counter / $TotalFiles) * 100)
        Write-Progress -Activity "Copiando archivos" -Status "$Counter de $TotalFiles archivos ($Percent%)" -PercentComplete $Percent
    }

    Write-Progress -Activity "Copiando archivos" -Completed

    Write-Host ""
    Write-Host "Archivos copiados exitosamente!" -ForegroundColor Green
    Write-Host "Total de archivos: $TotalFiles" -ForegroundColor Cyan

    # Mostrar resumen de lo copiado
    Write-Host ""
    Write-Host "Resumen de archivos copiados:" -ForegroundColor Yellow
    Get-ChildItem "${DriveName}:" -Recurse -File | Group-Object Extension | Sort-Object Count -Descending | Format-Table Name, Count -AutoSize

    # Listar estructura de directorios
    Write-Host ""
    Write-Host "Estructura de directorios copiada:" -ForegroundColor Yellow
    Get-ChildItem "${DriveName}:" -Recurse -Directory | ForEach-Object {
        $depth = ($_.FullName.Replace("${DriveName}:", "").Split("\").Length - 1)
        $indent = "  " * $depth
        Write-Host "$indent$($_.Name)" -ForegroundColor Gray
    }

} catch {
    Write-Host ""
    Write-Host "ERROR: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
} finally {
    # Desconectar unidad
    if (Get-PSDrive -Name $DriveName -ErrorAction SilentlyContinue) {
        Write-Host ""
        Write-Host "Desconectando unidad de red..." -ForegroundColor Yellow
        Remove-PSDrive -Name $DriveName -Force
        Write-Host "Desconectado" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Despliegue completado exitosamente" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Servidor: \\$ServerIP\$ShareName" -ForegroundColor Yellow
Write-Host ""
