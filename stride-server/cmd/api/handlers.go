package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/loosehose/stride/stride-server/internal"
	"github.com/loosehose/stride/stride-server/internal/deployment/common"
	"github.com/loosehose/stride/stride-server/internal/deployment/payloads"
	"github.com/loosehose/stride/stride-server/internal/deployment/phishing"
	"github.com/loosehose/stride/stride-server/internal/deployment/portforward"
	"github.com/loosehose/stride/stride-server/internal/deployment/redirector"
	"github.com/loosehose/stride/stride-server/internal/deployment/teamserver"
	"github.com/loosehose/stride/stride-server/internal/models"
)

// TeamserverSetupRequest defines the structure for a teamserver setup request.
type TeamserverSetupRequest struct {
	AgentName  string   `json:"agentName"`
	Software   []string `json:"software"`
	SSHKeyName string   `json:"sshKeyName"` // New field for SSH key name
}

// RedirectorSetupRequest defines the structure for a redirector setup request.
type RedirectorSetupRequest struct {
	RedirectorAgentIP string   `json:"redirectorAgentIP"`
	TeamserverAgentIP string   `json:"teamserverAgentIP"`
	Software          []string `json:"software"`
	Domain            string   `json:"domain"`
}

// PortForwardingSetupRequest defines the request structure for setting up port forwarding.
type PortForwardingSetupRequest struct {
	TeamserverAgentIP string `json:"teamserverAgent"`
	RedirectorAgentIP string `json:"redirectorAgent"`
	SourcePort        string `json:"sourcePort"`
	Protocol          string `json:"protocol"`
	DestinationPort   string `json:"destinationPort"`
}

// PhishingSetupRequest defines the structure for a phishing setup request.
type PhishingSetupRequest struct {
	AgentIP        string   `json:"agentIP"`
	RootDomain     string   `json:"rootDomain"`
	Subdomains     []string `json:"subdomains"`
	RootDomainBool string   `json:"rootDomainBool"`
	RedirectURL    string   `json:"redirectUrl"`
	FeedBool       string   `json:"feedBool"`
	RidReplacement string   `json:"ridReplacement"`
	BlacklistBool  string   `json:"blacklistBool"`
}

// setupPayloadHandler handles the creation and configuration of a new payload.
type PayloadSetupRequest struct {
	AgentIP       string `json:"agentIP"`
	ShellcodePath string `json:"shellcodePath"`
	Process       string `json:"process"`
	Method        string `json:"method"`
	UnhookNtdll   bool   `json:"unhookNtdll"`
	Syscall       string `json:"syscall"`
	DLL           bool   `json:"dll"` // Assuming you'll handle conversion in handler
	Outfile       string `json:"outfile"`
}

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "STRIDE up and running",
		Version: "1.0.0",
	}

	out, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

func (app *application) AllAgents(w http.ResponseWriter, r *http.Request) {
	// Extract the API key from request headers
	apiToken := r.Header.Get("X-Api-Key")
	if apiToken == "" {
		http.Error(w, "API key is required", http.StatusBadRequest)
		return
	}

	agents, err := internal.FetchDroplets(apiToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(agents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

func (app *application) AllDomains(w http.ResponseWriter, r *http.Request) {
	// Extract the API key from request headers
	apiToken := r.Header.Get("X-Api-Key")
	if apiToken == "" {
		http.Error(w, "API key is required", http.StatusBadRequest)
		return
	}

	domains, err := internal.FetchDomainsAndRecords(apiToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal and write the domain data as JSON
	out, err := json.Marshal(domains)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

// createSubdomainHandler processes requests to create subdomains.
func (app *application) createSubdomainHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the API key from request headers
	apiToken := r.Header.Get("X-Api-Key")
	if apiToken == "" {
		http.Error(w, "API key is required", http.StatusBadRequest)
		return
	}

	var request struct {
		DomainName    string `json:"domainName"`
		SubdomainName string `json:"subdomainName"`
		RecordType    string `json:"recordType"` // Typically "A" or "CNAME"
		Data          string `json:"data"`       // IP address for "A" record or hostname for "CNAME"
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Assuming API token is stored globally or retrieved securely
	if err := internal.CreateSubdomain(apiToken, request.DomainName, request.SubdomainName, request.RecordType, request.Data, app.wsManager); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create subdomain: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Subdomain created successfully")
}

// deleteSubdomainHandler handles requests to delete a subdomain.
func (app *application) deleteSubdomainHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the API key from request headers
	apiToken := r.Header.Get("X-API-Key")
	if apiToken == "" {
		http.Error(w, "API key is required", http.StatusBadRequest)
		return
	}

	// Extract domain name and subdomain name from URL path
	domainName := chi.URLParam(r, "domainName")
	subdomainName := chi.URLParam(r, "subdomainName")

	if err := internal.DeleteSubdomain(apiToken, domainName, subdomainName, app.wsManager); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete subdomain: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Subdomain deleted successfully")
}

// createAgentHandler handles the creation of a new agent, including setting up the SSH key.
func (app *application) createAgentHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the API key from request headers
	apiToken := r.Header.Get("X-Api-Key")
	if apiToken == "" {
		http.Error(w, "API key is required", http.StatusBadRequest)
		return
	}

	var request struct {
		Agent   string   `json:"agent"`
		SSHKeys []string `json:"ssh_keys"`
		Size    string   `json:"size"`
	}

	// Decode the incoming JSON to the struct
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create a new agent model from the decoded JSON
	agent := models.Agent{
		Name:    request.Agent,
		SSHKeys: request.SSHKeys,
		Size:    request.Size,
	}

	// Call the CreateAgent function with the constructed agent model
	if err := internal.CreateAgent(app.wsManager, r.Context(), &agent, apiToken); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create agent: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with the created agent details or a success message
	w.WriteHeader(http.StatusCreated)
}

// deleteAgentHandler handles requests to delete an agent (droplet).
func (app *application) deleteAgentHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the API key from request headers
	apiToken := r.Header.Get("X-Api-Key")
	if apiToken == "" {
		http.Error(w, "API key is required", http.StatusBadRequest)
		return
	}

	// Extract droplet ID from URL path or body. Here's an example using URL path with chi router.
	dropletIDStr := chi.URLParam(r, "dropletID")
	dropletID, err := strconv.Atoi(dropletIDStr)
	if err != nil {
		common.NewLogMessage(common.Error, "Invalid droplet ID", "ERROR", app.wsManager)
		http.Error(w, "Invalid droplet ID", http.StatusBadRequest)
		return
	}

	err = internal.DeleteAgent(app.wsManager, r.Context(), dropletID, apiToken)
	if err != nil {
		http.Error(w, "Failed to delete agent", http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusOK)
}

// listSSHKeysHandler handles requests to list all SSH keys in your DigitalOcean account.
func (app *application) listSSHKeysHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the API key from request headers
	apiToken := r.Header.Get("X-API-Key")
	if apiToken == "" {
		http.Error(w, "API key is required", http.StatusBadRequest)
		return
	}
	sshKeys, err := internal.ListSSHKeys(apiToken)
	if err != nil {
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to fetch SSH keys: %v", err), "", app.wsManager)
		http.Error(w, "Failed to fetch SSH keys", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sshKeys); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func getPrivateKeyPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".ssh", "id_rsa"), nil
}

// setupTeamserverHandler handles the creation of a new teamserver.
func (app *application) setupTeamserverHandler(w http.ResponseWriter, r *http.Request) {
	common.NewLogMessage(common.Exec, "Setting up teamserver", "EXEC", app.wsManager)
	var request struct {
		AgentIP  string   `json:"agentIP"`
		Software []string `json:"software"`
		ToastID  string   `json:"toastId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	privateKeyPath, err := getPrivateKeyPath()
	if err != nil {
		log.Printf("Failed to get private key path: %v", err)
		http.Error(w, "Failed to get private key path", http.StatusInternalServerError)
		return
	}

	// Loop through the software list and deploy accordingly
	for _, software := range request.Software {
		switch software {
		case "Sliver":
			if err := teamserver.DeploySliverTeamserver(app.wsManager, request.AgentIP, "22", "root", privateKeyPath); err != nil {
				common.NewLogMessage(common.Error, fmt.Sprintf("Failed to deploy Sliver: %v", err), "ERROR", app.wsManager)
				log.Printf("Failed to deploy Sliver: %v", err)
				http.Error(w, fmt.Sprintf("Error deploying Sliver: %v", err), http.StatusInternalServerError)
				return
			}
		case "Mythic":
			if err := teamserver.DeployMythicTeamserver(app.wsManager, request.AgentIP, "22", "root", privateKeyPath); err != nil {
				common.NewLogMessage(common.Error, fmt.Sprintf("Failed to deploy Mythic: %v", err), request.ToastID, app.wsManager)
				log.Printf("Failed to deploy Mythic: %v", err)
				http.Error(w, fmt.Sprintf("Error deploying Mythic: %v", err), http.StatusInternalServerError)
				return
			}
		case "HavocC2":
			if err := teamserver.DeployHavocC2Teamserver(app.wsManager, request.AgentIP, "22", "root", privateKeyPath); err != nil {
				common.NewLogMessage(common.Error, fmt.Sprintf("Failed to deploy HavocC2: %v", err), request.ToastID, app.wsManager)
				log.Printf("Failed to deploy HavocC2: %v", err)
				http.Error(w, fmt.Sprintf("Error deploying HavocC2: %v", err), http.StatusInternalServerError)
				return
			}
		default:
			// Handle other software options
		}
	}

	w.WriteHeader(http.StatusOK)
}

// setupRedirectorHandler handles the creation and configuration of a new redirector.
func (app *application) setupRedirectorHandler(w http.ResponseWriter, r *http.Request) {
	common.NewLogMessage(common.Exec, "Setting up redirector", "EXEC", app.wsManager)
	var req struct { // Directly use `req` here
		RedirectorAgentIP string   `json:"redirectorAgent"`
		TeamserverAgentIP string   `json:"teamserverAgent"`
		Software          []string `json:"software"`
		Domain            string   `json:"domain"`
		ToastID           string   `json:"toastId"`
	}

	// Decode the incoming JSON to the `req` struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	privateKeyPath, err := getPrivateKeyPath()
	if err != nil {
		log.Printf("Failed to get private key path: %v", err)
	}

	// Loop through the software list and deploy accordingly
	for _, software := range req.Software {
		switch software {
		case "Apache":
			if err := redirector.DeployApacheRedirector(app.wsManager, req.RedirectorAgentIP, req.TeamserverAgentIP, req.Domain, "22", "root", privateKeyPath, req.ToastID); err != nil { // Use `req` here
				// Handle error
				http.Error(w, fmt.Sprintf("Error deploying Apache: %v", err), http.StatusInternalServerError) // Corrected the error message to "Apache"
				return
			}
			// Handle other software options
		}
	}

	// Respond with success message
	w.WriteHeader(http.StatusOK)
}

// setupPortForwardingHandler handles requests to set up port forwarding.
func (app *application) setupPortForwardingHandler(w http.ResponseWriter, r *http.Request) {
	common.NewLogMessage(common.Exec, "Setting up port forwarding", "EXEC", app.wsManager)
	privateKeyPath, err := getPrivateKeyPath()
	if err != nil {
		log.Printf("Failed to get private key path: %v", err)
	}

	var request PortForwardingSetupRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call the function to set up port forwarding
	if err := portforward.DeployPortForwarding(app.wsManager, request.RedirectorAgentIP, request.TeamserverAgentIP, request.SourcePort, request.Protocol, request.DestinationPort, privateKeyPath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to set up port forwarding: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Port forwarding setup successfully")
}

// setupPhishingHandler handles the creation and configuration of a new phishing setup.
func (app *application) setupPhishingHandler(w http.ResponseWriter, r *http.Request) {
    apiToken := r.Header.Get("X-API-Key")
    if apiToken == "" {
        http.Error(w, "API key is required", http.StatusBadRequest)
        return
    }
    common.NewLogMessage(common.Exec, "Setting up phishing infrastructure", "EXEC", app.wsManager)
    var request PhishingSetupRequest
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    privateKeyPath, err := getPrivateKeyPath()
    if err != nil {
        log.Printf("Failed to get private key path: %v", err)
    }

    // Create a new DeploymentConfig struct
    deploymentConfig := phishing.DeploymentConfig{
        APIToken:        apiToken,
        AgentIP:         request.AgentIP,
        User:            "root",
        PrivateKeyPath:  privateKeyPath,
        RootDomain:      request.RootDomain,
        Subdomains:      request.Subdomains,
        RootDomainBool:  request.RootDomainBool,
        RedirectURL:     request.RedirectURL,
        FeedBool:        request.FeedBool,
        RIDReplacement:  request.RidReplacement,
        BlacklistBool:   request.BlacklistBool,
        CertsPath:       "/etc/letsencrypt/live/",
    }

    // Call DeployEvilGoPhish with the DeploymentConfig struct
    _, _, err = phishing.DeployEvilGoPhish(app.wsManager, deploymentConfig)
    if err != nil {
        common.NewLogMessage(common.Error, fmt.Sprintf("Failed to deploy EvilGoPhish: %v", err), "ERROR", app.wsManager)
        http.Error(w, fmt.Sprintf("Deployment failed: %v", err), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
}

func (app *application) setupPayloadHandler(w http.ResponseWriter, r *http.Request) {
	common.NewLogMessage(common.Exec, "Setting up payload", "EXEC", app.wsManager)
	var req PayloadSetupRequest

	// Direct decoding into the PayloadSetupRequest struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Construct the command line arguments based on the req object
	options := make([]string, 0)
	if req.Process != "" {
		options = append(options, fmt.Sprintf("-p %s", req.Process))
	}
	if req.Method != "" {
		options = append(options, fmt.Sprintf("-m %s", req.Method))
	}
	if req.UnhookNtdll {
		options = append(options, "-u")
	}

	if req.Syscall != "" {
		options = append(options, fmt.Sprintf("-sc %s", req.Syscall))
	}
	if req.DLL {
		options = append(options, "-d")
	}
	if req.Outfile != "" {
		options = append(options, fmt.Sprintf("-o %s", req.Outfile))
	}
	commandLineOptions := strings.Join(options, " ")

	// Call DeployShhhloader with the constructed command line options
	privateKeyPath, _ := getPrivateKeyPath() // Ensure proper error handling
	err := payloads.DeployShhhloader(app.wsManager, req.AgentIP, req.ShellcodePath, "root", privateKeyPath, commandLineOptions)
	if err != nil {
		log.Printf("Failed to deploy Shhhloader: %v", err)
		http.Error(w, fmt.Sprintf("Failed to deploy Shhhloader: %v", err), http.StatusInternalServerError)
		return
	}

	// Log success and respond to the client
	log.Println("Shhhloader deployed successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}
