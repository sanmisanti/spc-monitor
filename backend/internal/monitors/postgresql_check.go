package monitors

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saltacompra/monitor/internal/models"
)

// PostgreSQLCheckConfig contiene la configuración para el check de PostgreSQL
type PostgreSQLCheckConfig struct {
	Host           string
	Port           int
	User           string
	Password       string
	Database       string
	CheckID        string
	CheckName      string
	VPNCheckHost   string // Host para verificar VPN
	VPNTimeoutMs   int    // Timeout en ms para verificar VPN
}

// CheckVPNConnectivity verifica si hay conectividad con la red privada (VPN)
// Intenta hacer una conexión TCP simple al host especificado con timeout corto
func CheckVPNConnectivity(host string, port int, timeoutMs int) bool {
	timeout := time.Duration(timeoutMs) * time.Millisecond
	address := fmt.Sprintf("%s:%d", host, port)

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}

// CheckPostgreSQL verifica el estado de la conexión a PostgreSQL
// Primero verifica si hay conectividad VPN antes de intentar conectar
func CheckPostgreSQL(config PostgreSQLCheckConfig) models.Check {
	check := models.Check{
		ID:        config.CheckID,
		Type:      "database",
		Name:      config.CheckName,
		LastCheck: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Primero verificar si hay conectividad VPN
	vpnHost := config.VPNCheckHost
	if vpnHost == "" {
		vpnHost = config.Host // Si no se especifica, usar el mismo host de PostgreSQL
	}

	vpnTimeout := config.VPNTimeoutMs
	if vpnTimeout == 0 {
		vpnTimeout = 2000 // Default 2 segundos
	}

	hasVPN := CheckVPNConnectivity(vpnHost, config.Port, vpnTimeout)
	check.Metadata["vpn_check_host"] = vpnHost
	check.Metadata["vpn_available"] = hasVPN

	if !hasVPN {
		check.Status = "error"
		check.Message = fmt.Sprintf("No hay conectividad con la red privada (VPN). No se puede acceder a %s:%d", vpnHost, config.Port)
		check.Metadata["error_type"] = "vpn_unavailable"
		return check
	}

	// Si hay VPN, proceder con el check de PostgreSQL
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		config.User, config.Password, config.Host, config.Port, config.Database)

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		elapsed := time.Since(start).Milliseconds()
		check.ResponseTime = elapsed
		check.Status = "error"
		check.Message = "Error al conectar a PostgreSQL: " + err.Error()
		check.Metadata["error_type"] = "connection_failed"
		return check
	}
	defer conn.Close(ctx)

	// Ejecutar query para contar usuarios
	var userCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM users.usuarios").Scan(&userCount)
	elapsed := time.Since(start).Milliseconds()
	check.ResponseTime = elapsed

	if err != nil {
		check.Status = "error"
		check.Message = "Error al ejecutar query de conteo de usuarios: " + err.Error()
		check.Metadata["error_type"] = "query_failed"
		return check
	}

	// Todo OK
	check.Status = "ok"
	check.Message = fmt.Sprintf("Conexión PostgreSQL exitosa. %d usuarios registrados (%dms)", userCount, elapsed)
	check.Metadata["database"] = config.Database
	check.Metadata["server_version"] = conn.PgConn().ParameterStatus("server_version")
	check.Metadata["user_count"] = userCount

	return check
}
