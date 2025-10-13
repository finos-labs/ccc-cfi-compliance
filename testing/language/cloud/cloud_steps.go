package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/cucumber/godog"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/factory"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/generic"
)

// Connection represents a network connection with state and I/O
type Connection struct {
	State      string    // "open" or "closed"
	Input      io.Writer // Stream to write data to the connection
	Output     string    // Buffer containing all the data received from the connection so far
	cmd        *exec.Cmd // The underlying command process
	outputBuf  *bytes.Buffer
	stateMu    sync.Mutex    // Protects State field
	mu         sync.Mutex    // Protects Output field
	stopReader chan struct{} // Channel to signal the reader goroutine to stop
}

// Close terminates the connection and kills the underlying process
func (c *Connection) Close() {
	c.stateMu.Lock()
	c.State = "closed"
	c.stateMu.Unlock()

	// Signal the reader goroutine to stop
	if c.stopReader != nil {
		close(c.stopReader)
	}

	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}
}

// GetState returns the current connection state (thread-safe)
func (c *Connection) GetState() string {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	return c.State
}

// startOutputReader starts a goroutine that continuously reads from stdout and appends to Output
func (c *Connection) startOutputReader(reader io.Reader) {
	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-c.stopReader:
				return
			default:
				n, err := reader.Read(buf)
				if n > 0 {
					c.mu.Lock()
					c.outputBuf.Write(buf[:n])
					c.Output = c.outputBuf.String()
					c.mu.Unlock()
				}
				if err != nil {
					return
				}
			}
		}
	}()
}

// CloudWorld extends PropsWorld with cloud-specific functionality
type CloudWorld struct {
	*generic.PropsWorld
	mu sync.RWMutex
}

// NewCloudWorld creates a new CloudWorld instance
func NewCloudWorld() *CloudWorld {
	return &CloudWorld{
		PropsWorld: generic.NewPropsWorld(),
	}
}

// RegisterSteps registers all cloud-specific step definitions
func (cw *CloudWorld) RegisterSteps(ctx *godog.ScenarioContext) {
	// Register generic steps first
	cw.PropsWorld.RegisterSteps(ctx)

	// Cloud-specific steps matching Cucumber-Cloud-Language.md

	// OpenSSL connections
	ctx.Step(`^an openssl s_client request to "([^"]*)" on "([^"]*)" protocol "([^"]*)"$`, cw.opensslClientRequestWithProtocol)
	ctx.Step(`^an openssl s_client request using "([^"]*)" to "([^"]*)" on "([^"]*)" protocol "([^"]*)"$`, func(tlsVersion, port, host, protocol string) error {
		return cw.opensslClientRequestWithTLSAndProtocol(tlsVersion, port, host, protocol)
	})

	// Plain client connections
	ctx.Step(`^a client connects to "([^"]*)" with protocol "([^"]*)" on port "([^"]*)"$`, cw.clientConnectsWithProtocol)

	// Connection operations
	ctx.Step(`^I transmit "([^"]*)" to "([^"]*)"$`, cw.transmitToConnection)
	ctx.Step(`^I close connection "([^"]*)"$`, cw.closeConnection)
	ctx.Step(`^"([^"]*)" state is (open|closed)$`, cw.checkConnectionState)

	// SSL Support reports
	ctx.Step(`^"([^"]*)" contains details of SSL Support type "([^"]*)" for "([^"]*)" on port "([^"]*)"$`, cw.getSSLSupportReport)
	ctx.Step(`^"([^"]*)" contains details of SSL Support type "([^"]*)" for "([^"]*)" on port "([^"]*)" with STARTTLS$`, cw.getSSLSupportReportWithSTARTTLS)

	// Cloud API steps
	ctx.Step(`^a cloud api for "([^"]*)" in "([^"]*)"$`, cw.aCloudAPIForProviderIn)
}

// opensslClientRequest creates an OpenSSL s_client connection with optional TLS version
func (cw *CloudWorld) opensslClientRequest(tlsVersion, port, hostName, protocol string) error {
	tlsVersionResolved := cw.HandleResolve(tlsVersion)
	portResolved := cw.HandleResolve(port)
	hostResolved := cw.HandleResolve(hostName)
	protocolResolved := cw.HandleResolve(protocol)

	// Build openssl s_client command
	args := []string{"s_client", "-connect", fmt.Sprintf("%v:%v", hostResolved, portResolved), "-connect_timeout", "5"}

	// Add TLS version if specified
	if tlsVersionResolved != nil && fmt.Sprintf("%v", tlsVersionResolved) != "" {
		args = append(args, "-"+fmt.Sprintf("%v", tlsVersionResolved))
	}

	// Add STARTTLS if protocol is specified
	if protocolResolved != nil && fmt.Sprintf("%v", protocolResolved) != "" {
		args = append(args, "-starttls", fmt.Sprintf("%v", protocolResolved))
	}

	cmd := exec.Command("openssl", args...)

	// Create buffers for I/O
	inputBuffer := &bytes.Buffer{}
	outputBuffer := &bytes.Buffer{}

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	// Get stderr pipe
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	// Connect command's stdin to our input buffer
	cmd.Stdin = inputBuffer

	// Debug: Print the command being executed
	fmt.Printf("DEBUG: Executing: openssl %v\n", strings.Join(args, " "))

	// Start the command
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	// Create Connection object
	conn := &Connection{
		State:      "open",
		Input:      inputBuffer,
		Output:     "",
		cmd:        cmd,
		outputBuf:  outputBuffer,
		stopReader: make(chan struct{}),
	}

	// Start goroutines to read stdout and stderr
	conn.startOutputReader(stdout)
	conn.startOutputReader(stderr)

	// Monitor the command and set state to closed when it exits
	go func() {
		cmd.Wait()
		conn.stateMu.Lock()
		conn.State = "closed"
		conn.stateMu.Unlock()
		fmt.Printf("DEBUG: Command exited, connection state set to closed\n")
	}()

	cw.Props["result"] = conn
	fmt.Printf("DEBUG: Created connection with State=%v, stored in result\n", conn.State)
	return nil
}

// opensslClientRequestWithProtocol creates an OpenSSL s_client connection
func (cw *CloudWorld) opensslClientRequestWithProtocol(port, hostName, protocol string) error {
	return cw.opensslClientRequest("", port, hostName, protocol)
}

// opensslClientRequestWithTLSAndProtocol creates an OpenSSL s_client connection with specific TLS version
func (cw *CloudWorld) opensslClientRequestWithTLSAndProtocol(tlsVersion, port, hostName, protocol string) error {
	return cw.opensslClientRequest(tlsVersion, port, hostName, protocol)
}

// clientConnectsWithProtocol establishes a plain client connection to a host with a specific protocol
func (cw *CloudWorld) clientConnectsWithProtocol(hostName, protocol, port string) error {
	return cw.opensslClientRequest("", port, hostName, "")
}

// transmitToConnection sends data to a connection's input field
func (cw *CloudWorld) transmitToConnection(data, connectionInputPath string) error {
	dataResolved := cw.HandleResolve(data)

	if dataResolved == nil {
		return fmt.Errorf("data %s not found", data)
	}

	// The connectionInputPath should be something like "{connection.input}"
	// We need to extract the connection variable and set its Input field
	// The generic HandleResolve will handle the field access
	// For now, we'll use a simplified approach and directly set the input

	// Extract connection name from path like "{connection.input}"
	// This is handled by the generic PropsWorld field resolution
	inputStr := fmt.Sprintf("%v", dataResolved)

	// Store the transmission - in real implementation this would send over socket
	cw.Props["result"] = fmt.Sprintf("Transmitted: %v", inputStr)
	return nil
}

// closeConnection closes an established connection
func (cw *CloudWorld) closeConnection(connectionName string) error {
	// HandleResolve will resolve "{connection}" to the actual Connection object
	connInterface := cw.HandleResolve(connectionName)
	if connInterface == nil {
		return fmt.Errorf("connection %s not found", connectionName)
	}

	// Type assert to Connection
	if conn, ok := connInterface.(*Connection); ok {
		conn.Close()
	} else {
		return fmt.Errorf("connection %s is not a valid Connection object", connectionName)
	}

	return nil
}

// checkConnectionState verifies that a connection has the expected state
func (cw *CloudWorld) checkConnectionState(connectionName, expectedState string) error {
	// HandleResolve will resolve "{connection}" to the actual Connection object
	connInterface := cw.HandleResolve(connectionName)
	if connInterface == nil {
		return fmt.Errorf("connection %s not found", connectionName)
	}

	// Type assert to Connection
	conn, ok := connInterface.(*Connection)
	if !ok {
		return fmt.Errorf("connection %s is not a valid Connection object", connectionName)
	}

	// Thread-safe state access
	currentState := conn.GetState()
	if currentState != expectedState {
		return fmt.Errorf("connection %s state is %s, expected %s", connectionName, currentState, expectedState)
	}

	return nil
}

// runTestSSL is a helper function to run testssl.sh and return JSON report
func (cw *CloudWorld) runTestSSL(reportName, testType, hostName, port string, useSTARTTLS bool) error {
	reportNameResolved := cw.HandleResolve(reportName)
	testTypeResolved := cw.HandleResolve(testType)
	hostResolved := cw.HandleResolve(hostName)
	portResolved := cw.HandleResolve(port)

	// Create temporary file for JSON output
	tempFile := fmt.Sprintf("/tmp/testssl_%v_%v_%v", testTypeResolved, hostResolved, portResolved)
	if useSTARTTLS {
		tempFile += "_starttls"
	}
	tempFile += ".json"

	// Build testssl.sh command
	// Get the directory where this Go file is located
	_, filename, _, _ := runtime.Caller(0)
	cloudDir := filepath.Dir(filename)
	testsslPath := filepath.Join(cloudDir, "testssl.sh")

	args := []string{testsslPath, "--" + fmt.Sprintf("%v", testTypeResolved)}

	if useSTARTTLS {
		// Determine STARTTLS protocol from port
		protocol := cw.HandleResolve("{protocol}")
		if protocol == nil {
			protocol = "smtp" // default
		}
		args = append(args, "-t", fmt.Sprintf("%v", protocol))
	}

	args = append(args, "--jsonfile", tempFile, fmt.Sprintf("%v:%v", hostResolved, portResolved))

	// Remove the temporary JSON file if it exists from a previous run
	os.Remove(tempFile)

	cmd := exec.Command("bash", args...)

	// Debug: Print the command being executed
	fmt.Printf("DEBUG: Executing: bash %v\n", strings.Join(args, " "))

	_, err := cmd.CombinedOutput()
	if err != nil {
		// testssl.sh might return non-zero exit code even on success
		// Continue to try reading the JSON file
	}

	// Read and parse JSON output
	jsonData, err := os.ReadFile(tempFile)
	if err != nil {
		return fmt.Errorf("failed to read testssl.sh output: %v", err)
	}

	var report interface{}
	if err := json.Unmarshal(jsonData, &report); err != nil {
		return fmt.Errorf("failed to parse testssl.sh JSON: %v", err)
	}

	cw.Props[fmt.Sprintf("%v", reportNameResolved)] = report

	// Attach the JSON report for viewing in test results
	attachmentName := fmt.Sprintf("testssl_%v_%v_%v.json", testTypeResolved, hostResolved, portResolved)
	if useSTARTTLS {
		attachmentName = fmt.Sprintf("testssl_%v_%v_%v_starttls.json", testTypeResolved, hostResolved, portResolved)
	}
	cw.Attach(attachmentName, "application/json", jsonData)

	return nil
}

// getSSLSupportReport runs testssl.sh and returns JSON report
func (cw *CloudWorld) getSSLSupportReport(reportName, testType, hostName, port string) error {
	return cw.runTestSSL(reportName, testType, hostName, port, false)
}

// getSSLSupportReportWithSTARTTLS runs testssl.sh with STARTTLS support
func (cw *CloudWorld) getSSLSupportReportWithSTARTTLS(reportName, testType, hostName, port string) error {
	return cw.runTestSSL(reportName, testType, hostName, port, true)
}

// aCloudAPIForProviderIn initializes a cloud API factory for the specified provider
// Example: Given a cloud api for "aws" in "myAPI"
func (cw *CloudWorld) aCloudAPIForProviderIn(providerName string, apiName string) error {
	// Resolve the provider name (in case it's a variable reference)
	providerResolved := cw.HandleResolve(providerName)
	providerStr := fmt.Sprintf("%v", providerResolved)

	// Convert provider string to CloudProvider type
	var provider factory.CloudProvider
	switch providerStr {
	case "aws":
		provider = factory.ProviderAWS
	case "azure":
		provider = factory.ProviderAzure
	case "gcp":
		provider = factory.ProviderGCP
	default:
		return fmt.Errorf("unsupported cloud provider: %s (must be aws, azure, or gcp)", providerStr)
	}

	// Create the factory
	f, err := factory.NewFactory(provider)
	if err != nil {
		return fmt.Errorf("failed to create factory for provider %s: %w", providerStr, err)
	}

	// Store the factory in Props with the given API name
	cw.Props[apiName] = f

	return nil
}
