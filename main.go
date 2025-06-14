package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"

	"golang.org/x/net/proxy"
)

const (
	DefaultHealthcheckUrl         = "http://127.0.0.1:9081/healthz"
	DefaultLogLevel               = "INFO"
	DefaultPIDFile                = "/var/run/ggbridge.pid"
	DefaultPingFrequency          = 30
	DefaultServerIdleTimeout      = 30
	DefaultTunnelHealthPort       = 9081
	DefaultTunnelHealthRemotePort = 8081
	DefaultTunnelSocksPort        = 9180
	DefaultTunnelTlsPort          = 9443
	DefaultTunnelTlsRemotePort    = 8443
	DefaultTunnelWebPort          = 9080
	DefaultTunnelWebRemotePort    = 8443
)

// getEnv retrieves environment variables or default values.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func writePIDFile(pidFilePath string) error {
	// Get the current process ID
	pid := os.Getpid()

	// Open the file for writing (create if not exists, truncate if exists)
	file, err := os.OpenFile(pidFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open PID file: %w", err)
	}
	defer file.Close()

	// Write the PID to the file
	_, err = file.WriteString(strconv.Itoa(pid))
	if err != nil {
		return fmt.Errorf("failed to write PID to file: %w", err)
	}

	return nil
}

// performHealthCheck performs a health check using the specified URL
func performHealthCheck(healthCheckUrl string, proxyUrl string) (string, error) {

	// Create an HTTP client
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	if proxyUrl != "" {
		// Parse the proxy URL
		proxyUrl, err := url.Parse(proxyUrl)
		if err != nil {
			log.Fatalf("Invalid proxy address: %s", err)
		}
		// Create a SOCKS5 dialer
		dialer, err := proxy.SOCKS5("tcp", proxyUrl.Host, nil, proxy.Direct)
		if err != nil {
			return "", fmt.Errorf("error creating SOCKS5 dialer: %w", err)
		}

		// Create a custom HTTP transport
		transport := &http.Transport{
			Dial: func(network, healthCheckUrl string) (net.Conn, error) {
				return dialer.Dial(network, healthCheckUrl)
			},
		}
		// Update HTTP client transport
		client.Transport = transport
	}

	// Make the HTTP request
	resp, err := client.Get(healthCheckUrl)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	// Read and return the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	return string(body), nil
}

// buildClientCommand builds the command for client mode.
func buildClientCommand() []string {
	serverProtocol := getEnv("SERVER_PROTOCOL", "ws")
	serverAddress := os.Getenv("SERVER_ADDRESS")
	serverPort := os.Getenv("SERVER_PORT")
	serverPathPrefix := os.Getenv("SERVER_PATH_PREFIX")
	pingFrequency := getEnv("PING_FREQUENCY", strconv.Itoa(DefaultPingFrequency))
	connectionMinIdle := getEnv("CONNECTION_MIN_IDLE", "0")
	dnsResolver := os.Getenv("DNS_RESOLVER")
	tlsEnabled, err := strconv.ParseBool(getEnv("TLS_ENABLED", "false"))
	proxyProtocolEnabled, err := strconv.ParseBool(getEnv("PROXY_PROTOCOL_ENABLED", "true"))
	if err != nil {
		log.Fatalf("Invalid boolean for tlsEnabled: %s", err)
	}
	tlsVerifyCertificate, err := strconv.ParseBool(getEnv("TLS_VERIFY_CERTIFICATE", "false"))
	if err != nil {
		log.Fatalf("Invalid boolean for tlsVerifyCertificate: %s", err)
	}
	tunnelSocksEnabled, err := strconv.ParseBool(getEnv("TUNNEL_SOCKS_ENABLED", "false"))
	if err != nil {
		log.Fatalf("Invalid boolean for tunnelSocksEnabled: %s", err)
	}
	tunnelTlsEnabled, err := strconv.ParseBool(getEnv("TUNNEL_TLS_ENABLED", "false"))
	if err != nil {
		log.Fatalf("Invalid boolean for tunnelTlsEnabled: %s", err)
	}
	tunnelWebEnabled, err := strconv.ParseBool(getEnv("TUNNEL_WEB_ENABLED", "false"))
	if err != nil {
		log.Fatalf("Invalid boolean for tunnelWebEnabled: %s", err)
	}
	reverseTunnelSocksEnabled, err := strconv.ParseBool(getEnv("REVERSE_TUNNEL_SOCKS_ENABLED", "true"))
	if err != nil {
		log.Fatalf("Invalid boolean for reverseTunnelSocksEnabled: %s", err)
	}
	reverseTunnelTlsEnabled, err := strconv.ParseBool(getEnv("REVERSE_TUNNEL_TLS_ENABLED", "false"))
	if err != nil {
		log.Fatalf("Invalid boolean for reverseTunnelTlsEnabled: %s", err)
	}
	reverseTunnelWebEnabled, err := strconv.ParseBool(getEnv("REVERSE_TUNNEL_WEB_ENABLED", "false"))
	if err != nil {
		log.Fatalf("Invalid boolean for reverseTunnelWebEnabled: %s", err)
	}
	tunnelHealthPort := getEnv("TUNNEL_HEALTH_PORT", strconv.Itoa(DefaultTunnelHealthPort))
	tunnelHealthRemotePort := getEnv("TUNNEL_HEALTH_REMOTE_PORT", strconv.Itoa(DefaultTunnelHealthRemotePort))
	tunnelSocksPort := getEnv("TUNNEL_SOCKS_PORT", strconv.Itoa(DefaultTunnelSocksPort))
	tunnelTlsPort := getEnv("TUNNEL_TLS_PORT", strconv.Itoa(DefaultTunnelTlsPort))
	tunnelTlsRemotePort := getEnv("TUNNEL_TLS_REMOTE_PORT", strconv.Itoa(DefaultTunnelTlsRemotePort))
	tunnelWebPort := getEnv("TUNNEL_WEB_PORT", strconv.Itoa(DefaultTunnelWebPort))
	tunnelWebRemotePort := getEnv("TUNNEL_WEB_REMOTE_PORT", strconv.Itoa(DefaultTunnelWebRemotePort))

	if serverAddress == "" {
		fmt.Println("Error: SERVER_ADDRESS is mandatory")
		os.Exit(1)
	}

	if tlsEnabled {
		serverProtocol = "wss"
	}

	serverUrl := serverProtocol + "://" + serverAddress

	if serverPort != "" {
		serverUrl = serverUrl + ":" + serverPort
	}

	cmd := []string{
		"--log-lvl", getEnv("LOG_LEVEL", DefaultLogLevel),
		"client",
		serverUrl,
		"--websocket-ping-frequency-sec", pingFrequency,
		"--connection-min-idle", connectionMinIdle,
		// Healthcheck tunnel
		"--local-to-remote", fmt.Sprintf("tcp://0.0.0.0:%s:127.0.0.1:%s", tunnelHealthPort, tunnelHealthRemotePort),
		"--remote-to-local", fmt.Sprintf("tcp://0.0.0.0:%s:127.0.0.1:%s", tunnelHealthPort, tunnelHealthRemotePort),
	}

	// Use a specific prefix that will show up in the http path during the upgrade request
	if serverPathPrefix != "" {
		cmd = append(cmd, "--http-upgrade-path-prefix", serverPathPrefix)
	}

	// Add SSL flags if enabled
	if tlsEnabled {
		cmd = append(cmd, "--tls-certificate", "/etc/ggbridge/tls/client.crt")
		cmd = append(cmd, "--tls-private-key", "/etc/ggbridge/tls/client.key")
	}

	// Verify TLS Certificate
	if tlsVerifyCertificate {
		cmd = append(cmd, "--tls-verify-certificate")
	}

	// Add DNS resolver flag if set
	if dnsResolver != "" {
		cmd = append(cmd, "--dns-resolver", dnsResolver)
	}

	// Enables client to server proxy tunnel
	if tunnelSocksEnabled {
		cmd = append(cmd, "--local-to-remote", fmt.Sprintf("socks5://0.0.0.0:%s", tunnelSocksPort))
	}

	// Enables client to server tcp tunnel
	if tunnelTlsEnabled {
		target := fmt.Sprintf("tcp://0.0.0.0:%s:127.0.0.1:%s", tunnelTlsPort, tunnelTlsRemotePort)

		if proxyProtocolEnabled {
			target += "?proxy_protocol"
		}

		cmd = append(cmd, "--local-to-remote", target)
	}

	// Enables client to server web tunnel
	if tunnelWebEnabled {
		target := fmt.Sprintf("tcp://127.0.0.1:%s:127.0.0.1:%s", tunnelWebPort, tunnelWebRemotePort)

		if proxyProtocolEnabled {
			target += "?proxy_protocol"
		}

		cmd = append(cmd, "--local-to-remote", target)
	}

	// Enables server to client proxy tunnel
	if reverseTunnelSocksEnabled {
		cmd = append(cmd, "--remote-to-local", fmt.Sprintf("socks5://0.0.0.0:%s", tunnelSocksPort))
	}

	// Enables server to client tcp tunnel
	if reverseTunnelTlsEnabled {
		target := fmt.Sprintf("tcp://0.0.0.0:%s:127.0.0.1:%s", tunnelTlsPort, tunnelTlsRemotePort)

		if proxyProtocolEnabled {
			target += "?proxy_protocol"
		}

		cmd = append(cmd, "--remote-to-local", target)
	}

	// Enables server to client web tunnel
	if reverseTunnelWebEnabled {
		target := fmt.Sprintf("tcp://127.0.0.1:%s:127.0.0.1:%s", tunnelWebPort, tunnelWebRemotePort)

		if proxyProtocolEnabled {
			target += "?proxy_protocol"
		}

		cmd = append(cmd, "--remote-to-local", target)
	}

	return cmd
}

// buildServerCommand builds the command for server mode.
func buildServerCommand() []string {
	serverProtocol := getEnv("SERVER_PROTOCOL", "ws")
	serverListen := getEnv("SERVER_LISTEN", "0.0.0.0")
	serverPort := os.Getenv("SERVER_PORT")
	serverPathPrefix := os.Getenv("SERVER_PATH_PREFIX")
	serverIdleTimeout := getEnv("SERVER_IDLE_TIMEOUT", strconv.Itoa(DefaultServerIdleTimeout))
	pingFrequency := getEnv("PING_FREQUENCY", strconv.Itoa(DefaultPingFrequency))
	dnsResolver := os.Getenv("DNS_RESOLVER")
	restrictConfig := os.Getenv("RESTRICT_CONFIG")
	tlsEnabled, err := strconv.ParseBool(getEnv("TLS_ENABLED", "false"))
	if err != nil {
		log.Fatalf("Invalid boolean for tlsEnabled: %s", err)
	}

	if serverListen == "" {
		fmt.Println("Error: SERVER_LISTEN is mandatory")
		os.Exit(1)
	}

	if tlsEnabled {
		serverProtocol = "wss"
	}

	serverUrl := serverProtocol + "://" + serverListen

	if serverPort != "" {
		serverUrl = serverUrl + ":" + serverPort
	}

	cmd := []string{
		"--log-lvl", getEnv("LOG_LEVEL", DefaultLogLevel),
		"server",
		serverUrl,
		"--remote-to-local-server-idle-timeout-sec", serverIdleTimeout,
		"--websocket-ping-frequency-sec", pingFrequency,
	}

	// Server will only accept connection from if this specific path prefix is used during websocket upgrade.
	if serverPathPrefix != "" {
		cmd = append(cmd, "--restrict-http-upgrade-path-prefix", serverPathPrefix)
	}

	// Load restriction rules from config file
	if restrictConfig != "" {
		cmd = append(cmd, "--restrict-config", restrictConfig)
	}

	// Add SSL flags if enabled
	if tlsEnabled {
		cmd = append(cmd, "--tls-client-ca-certs", "/etc/ggbridge/tls/ca.crt")
		cmd = append(cmd, "--tls-certificate", "/etc/ggbridge/tls/server.crt")
		cmd = append(cmd, "--tls-private-key", "/etc/ggbridge/tls/server.key")
	}

	// Add DNS resolver flag if set
	if dnsResolver != "" {
		cmd = append(cmd, "--dns-resolver", dnsResolver)
	}

	return cmd
}

// runCommand runs a command and pipes its output
func runCommand(name string, args []string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error running command '%s': %v", name, err)
	}
}

// runNginx starts the embedded NGINX process
func runNginx() {
	runCommand("nginx", []string{"-c", "/etc/ggbridge/nginx.conf", "-e", "/dev/stderr"})
}

// runClient handles the client subcommand logic
func runClient(pidFile string) {
	err := writePIDFile(pidFile)
	if err != nil {
		log.Fatalf("Error writing PID file: %v", err)
	}

	if getEnv("NGINX_EMBEDDED", "true") == "true" {
		runNginx()
	}

	cmd := buildClientCommand()
	runCommand("wstunnel", cmd)
}

// runServer handles the server subcommand logic
func runServer(pidFile string) {
	err := writePIDFile(pidFile)
	if err != nil {
		log.Fatalf("Error writing PID file: %v", err)
	}

	if getEnv("NGINX_EMBEDDED", "true") == "true" {
		runNginx()
	}

	cmd := buildServerCommand()
	runCommand("wstunnel", cmd)
}

// runHealthCheck handles the healthcheck subcommand
func runHealthCheck(healthCheckUrl string, proxyUrl string, pidFile string, gracePeriod int) {
	if gracePeriod > 0 && pidFile != "" {
		// Check if the PID file exists
		fileInfo, err := os.Stat(pidFile)
		var startTime time.Time
		if err == nil {
			startTime = fileInfo.ModTime()
		} else {
			startTime = time.Now()
		}

		// Calculate elapsed time
		elapsedTime := time.Since(startTime).Seconds()
		// Check if within grace period
		if int(elapsedTime) < gracePeriod {
			log.Print("Within grace period, skipping healthcheck errors")
			os.Exit(0)
		}
	}

	result, err := performHealthCheck(healthCheckUrl, proxyUrl)
	if err != nil {
		log.Fatalf("Healthcheck failed: %v", err)
	}
	log.Printf("Healthcheck passed: %s", result)
}

func main() {
	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)
	clientPidFile := clientCmd.String("pid-file", DefaultPIDFile, "PID file path")

	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	serverPidFile := serverCmd.String("pid-file", DefaultPIDFile, "PID file path")

	healthCheckCmd := flag.NewFlagSet("healthcheck", flag.ExitOnError)
	healthCheckPidFile := healthCheckCmd.String("pid-file", DefaultPIDFile, "PID file path")
	healthCheckGracePeriod := healthCheckCmd.Int("grace-period", 0, "Grace period in seconds to ignore healthcheck failures since pod startup")
	healthCheckProxyUrl := healthCheckCmd.String("proxy", "", "Proxy address")

	if len(os.Args) < 2 {
		log.Fatal("Error: Missing subcommand 'server', 'client', or 'healthcheck'")
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "client":
		clientCmd.Parse(os.Args[2:])
		runClient(*clientPidFile)
	case "server":
		serverCmd.Parse(os.Args[2:])
		runServer(*serverPidFile)
	case "healthcheck":
		healthCheckCmd.Parse(os.Args[2:])
		healthCheckUrl := DefaultHealthcheckUrl
		if len(healthCheckCmd.Args()) > 0 {
			healthCheckUrl = healthCheckCmd.Arg(0)
		}
		runHealthCheck(healthCheckUrl, *healthCheckProxyUrl, *healthCheckPidFile, *healthCheckGracePeriod)
	default:
		log.Fatalf("Error: Unknown subcommand '%s'", subcommand)
	}
}
