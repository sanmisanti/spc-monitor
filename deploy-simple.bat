@echo off
REM Script de Despliegue Simple - Llamada al script PowerShell
REM Ejecuta el script PowerShell con los parámetros predeterminados

echo.
echo ========================================
echo   Ejecutando Despliegue...
echo ========================================
echo.

powershell -ExecutionPolicy Bypass -File "%~dp0deploy.ps1"

echo.
pause
