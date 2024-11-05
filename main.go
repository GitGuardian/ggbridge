package main

import (
	"fmt"
	"os"
	"os/exec"
)

const RemoteProxyProtocol = "socks5"
const RemoteProxyListen = "0.0.0.0"
const RemoteProxyPort = 1080

// getEnv retrieves environment variables or default values.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
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

func run(args []string) {
	cmd := exec.Command("wstunnel", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Error running ggbridge:", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Error: Missing argument 'server' or 'client'")
		os.Exit(1)
	}

	cmd := []string{
		"--log-lvl",
		getEnv("LOG_LEVEL", "INFO"),
	}

	mode := os.Args[1]

	if mode == "client" {
		cmd = append(cmd, buildClientCommand()...)
	} else if mode == "server" {
		cmd = append(cmd, buildServerCommand()...)
	} else {
		fmt.Println("Error: You must run 'client' or 'server' mode")
		os.Exit(1)
	}
	run(cmd)
}
