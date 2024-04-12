package internal

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/loosehose/stride/stride-server/internal/deployment/common"
	"github.com/loosehose/stride/stride-server/internal/models"
	"github.com/loosehose/stride/stride-server/logging"

	"github.com/digitalocean/godo"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type tokenSource struct {
	AccessToken string
}

// Domain and DomainRecord structs here match the structure of the data you want to return
type Domain struct {
	Name    string         `json:"name"`
	Records []DomainRecord `json:"records"`
}

type DomainRecord struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Data string `json:"data"`
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func init() {
	logging.InitLogger()
}

func FetchDroplets(apiToken string) ([]models.Agent, error) {
	tokenSource := &tokenSource{AccessToken: apiToken}
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	client := godo.NewClient(oauthClient)

	// Create a context
	ctx := context.TODO()

	// List all droplets
	droplets, _, err := client.Droplets.List(ctx, &godo.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch droplets")
		return nil, err
	}

	var agents []models.Agent
	for _, droplet := range droplets {
		agent := models.Agent{
			ID:      strconv.Itoa(droplet.ID), // Make sure your Agent struct has an ID field of type string
			Name:    droplet.Name,
			IP:      extractIP(droplet.Networks),
			Created: droplet.Created,
			Tags:    droplet.Tags,
		}
		agents = append(agents, agent)
	}

	log.Debug().Int("count", len(agents)).Msg("Fetched droplets")
	return agents, nil
}

// extractIP is a helper function to extract the droplet's public IPv4 address
func extractIP(networks *godo.Networks) string {
	for _, v4 := range networks.V4 {
		if v4.Type == "public" {
			return v4.IPAddress
		}
	}
	return "" // Return an empty string if no public IP was found
}

func FetchDomainsAndRecords(apiToken string) ([]Domain, error) {
	client := godo.NewClient(oauth2.NewClient(context.Background(), &tokenSource{AccessToken: apiToken}))
	domains, _, err := client.Domains.List(context.Background(), &godo.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch domains")
		return nil, err
	}

	var result []Domain
	for _, domain := range domains {
		records, _, err := client.Domains.Records(context.Background(), domain.Name, &godo.ListOptions{})
		if err != nil {
			log.Warn().Err(err).Str("domain", domain.Name).Msg("Failed to fetch domain records")
			continue
		}

		var domainRecords []DomainRecord
		for _, record := range records {
			if record.Type == "A" || record.Type == "CNAME" {
				name := record.Name
				if name != "@" && name != "" {
					name += "." + domain.Name
				} else {
					name = domain.Name
				}

				domainRecord := DomainRecord{
					Type: record.Type,
					Name: name,
					Data: record.Data,
				}
				domainRecords = append(domainRecords, domainRecord)
			}
		}

		result = append(result, Domain{Name: domain.Name, Records: domainRecords})
	}

	log.Info().Int("count", len(result)).Msg("Fetched domains and records")
	return result, nil
}

func CreateSubdomain(apiToken string, domainName, subdomainName, recordType, data string, wsm *common.WebSocketManager) error {
	common.NewLogMessage(common.Exec, fmt.Sprintf("Creating subdomain: %s.%s", subdomainName, domainName), "EXEC", wsm)
	client := godo.NewClient(oauth2.NewClient(context.Background(), &tokenSource{AccessToken: apiToken}))

	createRequest := &godo.DomainRecordEditRequest{
		Type: recordType,
		Name: subdomainName,
		Data: data,
	}

	_, _, err := client.Domains.CreateRecord(context.Background(), domainName, createRequest)
	if err != nil {
		log.Error().Err(err).Str("subdomain", subdomainName).Str("domain", domainName).Msg("Failed to create subdomain")
		return err
	}
	common.NewLogMessage(common.Success, fmt.Sprintf("Subdomain created successfully: %s.%s", subdomainName, domainName), "SUCCESS", wsm)
	log.Info().Str("subdomain", subdomainName).Str("domain", domainName).Msg("Subdomain created successfully")
	return nil
}

func DeleteSubdomain(apiToken, domainName, subdomainName string, wsm *common.WebSocketManager) error {
	common.NewLogMessage(common.Exec, fmt.Sprintf("Deleting subdomain: %s.%s", subdomainName, domainName), "EXEC", wsm)
	client := godo.NewClient(oauth2.NewClient(context.Background(), &tokenSource{AccessToken: apiToken}))

	records, _, err := client.Domains.Records(context.Background(), domainName, &godo.ListOptions{})
	if err != nil {
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to fetch domain records: %v", err), "ERROR", wsm)
		log.Error().Err(err).Str("domain", domainName).Msg("Failed to fetch domain records")
		return err
	}

	for _, record := range records {
		if record.Name == subdomainName && (record.Type == "A" || record.Type == "CNAME") {
			_, err := client.Domains.DeleteRecord(context.Background(), domainName, record.ID)
			if err != nil {
				common.NewLogMessage(common.Error, fmt.Sprintf("Failed to delete subdomain: %v", err), "ERROR", wsm)
				log.Error().Err(err).Str("subdomain", subdomainName).Str("domain", domainName).Msg("Failed to delete subdomain")
				return err
			}
			common.NewLogMessage(common.Success, fmt.Sprintf("Subdomain deleted successfully: %s.%s", subdomainName, domainName), "SUCCESS", wsm)
			log.Info().Str("subdomain", subdomainName).Str("domain", domainName).Msg("Subdomain deleted successfully")
			return nil
		}
	}

	common.NewLogMessage(common.Error, fmt.Sprintf("Subdomain not found: %s.%s", subdomainName, domainName), "ERROR", wsm)
	log.Warn().Str("subdomain", subdomainName).Str("domain", domainName).Msg("Subdomain not found")
	return fmt.Errorf("subdomain not found: %s.%s", subdomainName, domainName)
}

func CreateAgent(wsm *common.WebSocketManager, ctx context.Context, agent *models.Agent, apiToken string) error {
	log.Info().Str("name", agent.Name).Msg("Creating droplet")
	tokenSource := &tokenSource{AccessToken: apiToken}
	oauthClient := oauth2.NewClient(ctx, tokenSource)
	client := godo.NewClient(oauthClient)

	sshKeys := make([]godo.DropletCreateSSHKey, len(agent.SSHKeys))
	for i, keyID := range agent.SSHKeys {
		keyIDInt, err := strconv.Atoi(keyID)
		if err != nil {
			common.NewLogMessage(common.Error, fmt.Sprintf("Failed to convert SSH key ID: %v", err), "ERROR", wsm)
			log.Error().Err(err).Str("keyID", keyID).Msg("Failed to convert SSH key ID")
			return err
		}
		sshKeys[i] = godo.DropletCreateSSHKey{ID: keyIDInt}
	}

	createRequest := &godo.DropletCreateRequest{
		Name:    agent.Name,
		Region:  "nyc3",
		Size:    agent.Size,
		Image:   godo.DropletCreateImage{Slug: "ubuntu-20-04-x64"},
		SSHKeys: sshKeys,
		Tags:    agent.Software,
	}

	common.NewLogMessage(common.Exec, fmt.Sprintf("Creating droplet: %s", agent.Name), "EXEC", wsm)

	droplet, _, err := client.Droplets.Create(ctx, createRequest)
	if err != nil {
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to create droplet: %v", err), "ERROR", wsm)
		log.Error().Err(err).Str("name", agent.Name).Msg("Failed to create droplet")
		return err
	}

	agent.IP, err = waitForDropletIP(client, ctx, droplet.ID, wsm)
	if err != nil {
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to retrieve droplet IP: %v", err), "ERROR", wsm)
		log.Error().Err(err).Int("dropletID", droplet.ID).Msg("Failed to retrieve droplet IP")
		return err
	}
	time.Sleep(5 * time.Second)
	common.NewLogMessage(common.Success, fmt.Sprintf("Droplet %s successfully created with IP %s", agent.Name, agent.IP), "SUCCESS", wsm)
	log.Info().Str("name", agent.Name).Str("ip", agent.IP).Msg("Droplet successfully created")

	return nil
}

func waitForDropletIP(client *godo.Client, ctx context.Context, dropletID int, wsm *common.WebSocketManager) (string, error) {
	maxAttempts := 15
	for attempt := 0; attempt < maxAttempts; attempt++ {
		time.Sleep(10 * time.Second)

		droplet, _, err := client.Droplets.Get(ctx, dropletID)
		if err != nil {
			log.Warn().Err(err).Int("DropletID", dropletID).Msg("Failed to get droplet")
			continue
		}

		if len(droplet.Networks.V4) > 0 {
			log.Info().Str("IP", droplet.Networks.V4[0].IPAddress).Int("DropletID", dropletID).Msg("Droplet IP address retrieved")
			return droplet.Networks.V4[0].IPAddress, nil
		}

		common.NewLogMessage(common.Exec, fmt.Sprintf("Waiting for agent %d to get an IP address (attempt %d/%d)", dropletID, attempt+1, maxAttempts), "EXEC", wsm)
		log.Debug().Int("DropletID", dropletID).Int("Attempt", attempt+1).Int("MaxAttempts", maxAttempts).Msg("Waiting for droplet IP address")
	}

	return "", fmt.Errorf("no IPv4 address found for droplet %d after %d attempts", dropletID, maxAttempts)
}

func DeleteAgent(wsm *common.WebSocketManager, ctx context.Context, dropletID int, apiToken string) error {
	common.NewLogMessage(common.Exec, fmt.Sprintf("Deleting droplet %d", dropletID), "EXEC", wsm)
	client := godo.NewClient(oauth2.NewClient(ctx, &tokenSource{AccessToken: apiToken}))

	_, err := client.Droplets.Delete(ctx, dropletID)
	if err != nil {
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to delete droplet: %v", err), "ERROR", wsm)
		log.Error().Err(err).Int("dropletID", dropletID).Msg("Failed to delete droplet")
		return err
	}

	common.NewLogMessage(common.Success, fmt.Sprintf("Droplet %d successfully deleted", dropletID), "SUCCESS", wsm)
	log.Info().Int("dropletID", dropletID).Msg("Droplet successfully deleted")
	return nil
}

func ListSSHKeys(apiToken string) ([]models.SSHKey, error) {
	tokenSource := &tokenSource{AccessToken: apiToken}
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	client := godo.NewClient(oauthClient)

	keys, _, err := client.Keys.List(context.TODO(), &godo.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to list SSH keys")
		return nil, err
	}

	var sshKeys []models.SSHKey
	for _, key := range keys {
		sshKeys = append(sshKeys, models.SSHKey{
			ID:        strconv.Itoa(key.ID),
			Name:      key.Name,
			PublicKey: key.PublicKey,
		})
	}

	log.Debug().Int("count", len(sshKeys)).Msg("Listed SSH keys")
	return sshKeys, nil
}
