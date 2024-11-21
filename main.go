package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"golang.org/x/net/proxy"
)

const (
	RemoteProxyProtocol   = "socks5"
	RemoteProxyListen     = "0.0.0.0"
	RemoteProxyPort       = 1080
	DefaultLogLevel       = "INFO"
	DefaultHealthCheckURL = "http://127.0.0.1:8080/healthz"
	DefaultPIDFile        = "/var/run/ggbridge.pid"
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

// performHealthCheck performs a health check via SOCKS5 proxy
func performHealthCheck(url string) (string, error) {
	proxyAddr := fmt.Sprintf("127.0.0.1:%d", RemoteProxyPort)

	// Create a SOCKS5 dialer
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return "", fmt.Errorf("error creating SOCKS5 dialer: %w", err)
	}

	// Create a custom HTTP transport
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
	}

	// Create an HTTP client with the custom transport
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	// Make the HTTP request
	resp, err := client.Get(url)
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
	tlsEnabled := getEnv("TLS_ENABLED", "false")
	pingFrequency := getEnv("PING_FREQUENCY", "10")
	connectionMinIdle := getEnv("CONNECTION_MIN_IDLE", "0")
	dnsResolver := os.Getenv("DNS_RESOLVER")

	if serverAddress == "" {
		fmt.Println("Error: SERVER_ADDRESS is mandatory")
		os.Exit(1)
	}

	if tlsEnabled == "true" {
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
		"--remote-to-local", fmt.Sprintf("%s://%s:%d", RemoteProxyProtocol, RemoteProxyListen, RemoteProxyPort),
	}

	// Use a specific prefix that will show up in the http path during the upgrade request
	if serverPathPrefix != "" {
		cmd = append(cmd, "--http-upgrade-path-prefix", serverPathPrefix)
	}

	// Add SSL flags if enabled
	if tlsEnabled == "true" {
		cmd = append(cmd, "--tls-certificate", "/certs/client.crt")
		cmd = append(cmd, "--tls-private-key", "/certs/client.key")
	}

	// Add DNS resolver flag if set
	if dnsResolver != "" {
		cmd = append(cmd, "--dns-resolver", dnsResolver)
	}

	return cmd
}

// buildServerCommand builds the command for server mode.
func buildServerCommand() []string {
	serverProtocol := getEnv("SERVER_PROTOCOL", "ws")
	serverListen := getEnv("SERVER_LISTEN", "0.0.0.0")
	serverPort := os.Getenv("SERVER_PORT")
	serverPathPrefix := os.Getenv("SERVER_PATH_PREFIX")
	tlsEnabled := getEnv("TLS_ENABLED", "false")
	pingFrequency := getEnv("PING_FREQUENCY", "10")

	if serverListen == "" {
		fmt.Println("Error: SERVER_LISTEN is mandatory")
		os.Exit(1)
	}

	if tlsEnabled == "true" {
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
		"--websocket-ping-frequency-sec", pingFrequency,
		"--restrict-to",
		fmt.Sprintf("%s://%s:%d", RemoteProxyProtocol, RemoteProxyListen, RemoteProxyPort),
	}

	// Server will only accept connection from if this specific path prefix is used during websocket upgrade.
	if serverPathPrefix != "" {
		cmd = append(cmd, "--restrict-http-upgrade-path-prefix", serverPathPrefix)
	}

	// Add SSL flags if enabled
	if tlsEnabled == "true" {
		cmd = append(cmd, "--tls-client-ca-certs", "/certs/ca.crt")
		cmd = append(cmd, "--tls-certificate", "/certs/server.crt")
		cmd = append(cmd, "--tls-private-key", "/certs/server.key")
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
	runCommand("nginx", []string{"-c", "/etc/nginx/nginx.conf", "-e", "/dev/stderr"})
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
func runHealthCheck(pidFile string, gracePeriod int) {
	if gracePeriod > 0 {
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

	healthCheckURL := getEnv("HEALTHCHECK_URL", DefaultHealthCheckURL)
	result, err := performHealthCheck(healthCheckURL)
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
		runHealthCheck(*healthCheckPidFile, *healthCheckGracePeriod)
	default:
		log.Fatalf("Error: Unknown subcommand '%s'", subcommand)
	}
}
