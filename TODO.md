# TODOs - Sistema de Monitoreo

## Descubiertos por Claude

### Backend
- [ ] Implementar scheduler para jobs automáticos periódicos
- [ ] Agregar WebSocket para updates en tiempo real
- [ ] Implementar autenticación y seguridad
- [ ] Agregar checks para App.SaltaCompra (PostgreSQL)
- [ ] Agregar checks para Google Apps Script
- [ ] Analizar y refactorizar handlers.go (extraer configuración de sistemas hardcodeada, separar concerns)

### Frontend
- [ ] Crear aplicación React + Vite completa
- [ ] Implementar dashboard principal
- [ ] Agregar visualización de métricas (Recharts o similar)

### Otros
- [ ] Implementar integración con DevOps (revisar mensajes sin responder, monitorear pipelines, alertas de builds fallidos)

---

## Descubiertos por Santi

### Backend
- [x] Implementar check de expiración de dominio vía RDAP
- [x] Crear sistema "Infraestructura Compartida" para monitorear recursos compartidos
- [x] Implementar conteo dinámico de estados de mails (sent, unsent, failed, retrying)
- [x] Agregar umbral de detección de cola atascada (unsent)
- [x] Corregir tabla de correos (msdb.dbo.sysmail_mailitems)
- [x] Chequear estado de mails en preproducción
- [ ] Agregar complejidad al análisis de estado de mails (verificar por tipo de correo)
- [ ] Mostrar información detallada del último correo enviado (remitente, destinatario, asunto, last_mod_date, sent_date, sent_status)

### Frontend
- [ ]

### DevOps
- [ ]

### Otros
- [ ]

---

## Notas

Este archivo registra tareas pendientes y mejoras identificadas durante el desarrollo. No todas tienen la misma prioridad ni requieren resolución inmediata.
