package redirector

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/loosehose/stride/stride-server/internal/deployment/common"
	"github.com/loosehose/stride/stride-server/logging"
	"github.com/rs/zerolog/log"
)

func init() {
	logging.InitLogger()
}

func deployRedirector(wsm *common.WebSocketManager, agentIP, teamserverIP, domain, port, user, privateKeyPath, toastId string, commands []string) error {
	sshClient, err := common.NewSSHClient(agentIP, port, user, privateKeyPath)
	if err != nil {
		log.Error().Msgf("Failed to establish SSH connection to %s: %v", agentIP, err)
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to establish SSH connection to %s: %v", agentIP, err), "ERROR", wsm)
		return fmt.Errorf("error deploying Apache: %v", err)
	}

	// Execute the commands over SSH
	totalCommands := len(commands)
	for i, cmd := range commands {
		// Notify the front end about the current step
		common.NewLogMessage(common.Exec, fmt.Sprintf("Progress: %d/%d", i+1, totalCommands), "EXEC", wsm)

		log.Info().Msgf("Executing: %s", cmd)
		_, err = sshClient.ExecuteCommand(cmd)
		if err != nil {
			log.Error().Msgf("Failed to execute command: %v", err)
			common.NewLogMessage(common.Error, fmt.Sprintf("Failed to execute command: %v", err), "ERROR", wsm)
			return fmt.Errorf("failed to execute command: %v", err)
		}
	}

	teamServerPublicKey, err := RetrieveTeamserverPublicKey(teamserverIP, "22", user, privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to retrieve teamserver public key: %v", err)
	}

	// Append the teamserver's public key to the redirector's authorized_keys
	appendPublicKeyCmd := fmt.Sprintf("echo '%s' >> ~/.ssh/authorized_keys", teamServerPublicKey)
	if _, err := sshClient.ExecuteCommand(appendPublicKeyCmd); err != nil {
		return fmt.Errorf("failed to append teamserver public key to authorized_keys: %v", err)
	}
	log.Info().Msg("Appended teamserver public key to redirector's authorized_keys")

	log.Info().Msg("Redirector setup successful")
	common.NewLogMessage(common.Success, "Redirector setup successful", "INFO", wsm)
	return nil
}

// DeployApacheRedirector installs and configures Apache on a redirector agent.
func DeployApacheRedirector(wsm *common.WebSocketManager, agentIP, teamserverIP, domain, port, user, privateKeyPath, toastId string) error {
	// Get the absolute path of the installer.go file
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to determine the path of installer.go")
	}
	currentDir := filepath.Dir(currentFile)

	defaultSSLTemplate1Path := filepath.Join(currentDir, "1-default-ssl.conf.tmpl")
	defaultSSLTemplate2Path := filepath.Join(currentDir, "2-default-ssl.conf.tmpl")
	htaccessPath 			:= filepath.Join(currentDir, ".htaccess.tmpl")

	// Load the service file template
	defaultSSLTemplate1, err := common.LoadFile(defaultSSLTemplate1Path)
	if err != nil {
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to load 1-default-ssl.conf template: %v", err), "ERROR", wsm)
		return fmt.Errorf("failed to load 1-default-ssl.conf template: %v", err)
	}

	defaultSSLContent1 := strings.Replace(defaultSSLTemplate1, "{{DOMAIN}}", domain, -1)

	tempFile, err := ioutil.TempFile(os.TempDir(), "tmp-1-default-ssl.conf-")
	if err != nil {
		log.Fatal().Msg("Cannot create temporary file")
		return err
	}

	defer os.Remove(tempFile.Name())

	if _, err = tempFile.Write([]byte(defaultSSLContent1)); err != nil {
		log.Fatal().Msg("Failed to write to temporary file")
		return err
	}

	sshClient, err := common.NewSSHClient(agentIP, port, user, privateKeyPath)
	if err != nil {
		log.Error().Msgf("Failed to establish SSH connection to %s: %v", agentIP, err)
		common.NewLogMessage(common.Error, fmt.Sprintf("Error deploying Apache: %v", err), "ERROR", wsm)
		return fmt.Errorf("error deploying Apache: %v", err)
	}

	if err := sshClient.TransferFile(tempFile.Name(), "/tmp/1-default-ssl.conf"); err != nil {
		return err
	}
	log.Info().Msg("1-default-ssl.conf transferred")

	// Load the service file template
	defaultSSLTemplate2, err := common.LoadFile(defaultSSLTemplate2Path)
	if err != nil {
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to load 2-default-ssl.conf template: %v", err), "ERROR", wsm)
		return fmt.Errorf("failed to load 2-default-ssl.conf template: %v", err)
	}

	// Replace placeholder with actual redirector IP
	defaultSSLContent2 := strings.Replace(defaultSSLTemplate2, "{{DOMAIN}}", domain, -1)

	tempFile, err = ioutil.TempFile(os.TempDir(), "tmp-2-default-ssl.conf-")
	if err != nil {
		log.Fatal().Msg("Cannot create temporary file")
		return err
	}
	defer os.Remove(tempFile.Name())

	if _, err = tempFile.Write([]byte(defaultSSLContent2)); err != nil {
		log.Fatal().Msg("Failed to write to temporary file")
		return err
	}

	if err := sshClient.TransferFile(tempFile.Name(), "/tmp/2-default-ssl.conf"); err != nil {
		return err
	}
	log.Info().Msg("2-default-ssl.conf transferred")

	// Load the service file template
	htaccessTemplate, err := common.LoadFile(htaccessPath)
	if err != nil {
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to load htaccess template: %v", err), "ERROR", wsm)
		return fmt.Errorf("failed to load 2-default-ssl.conf template: %v", err)
	}

	// Replace placeholder with actual redirector IP
	htaccessContents := strings.Replace(htaccessTemplate, "{{TEAMSERVER_IP}}", teamserverIP, -1)

	tempFile, err = ioutil.TempFile(os.TempDir(), ".htaccess-")
	if err != nil {
		log.Fatal().Msg("Cannot create temporary file")
		return err
	}
	defer os.Remove(tempFile.Name())

	if _, err = tempFile.Write([]byte(htaccessContents)); err != nil {
		log.Fatal().Msg("Failed to write to temporary file")
		return err
	}
	if err := sshClient.TransferFile(tempFile.Name(), "/tmp/.htaccess"); err != nil {
		return err
	}

	log.Info().Msg(".htaccess transferred")

	commands := []string{
		"sudo apt update -y",
		"sudo apt install -y apache2 certbot python3-certbot-apache",
		"sudo a2enmod ssl rewrite proxy proxy_http",
		"sudo rm /etc/apache2/sites-enabled/000-default.conf",
		"mv /tmp/1-default-ssl.conf /etc/apache2/sites-enabled/default-ssl.conf",
		"sudo systemctl restart apache2",
		"certbot certonly -d " + domain + " --apache --server https://acme-v02.api.letsencrypt.org/directory --register-unsafely-without-email --agree-tos",
		"iptables -I INPUT -p tcp -m tcp --dport 443 -j ACCEPT",
		"iptables -t nat -A PREROUTING -p tcp --dport 443 -j DNAT --to-destination "+ teamserverIP +":443",
		"iptables -t nat -A POSTROUTING -j MASQUERADE",
		"iptables -I FORWARD -j ACCEPT",
		"iptables -P FORWARD ACCEPT",
		"sysctl net.ipv4.ip_forward=1",
		"mv /tmp/2-default-ssl.conf /etc/apache2/sites-enabled/default-ssl.conf",
		"mv /tmp/.htaccess /var/www/html/.htaccess",
		"sudo systemctl restart apache2",
	}

	return deployRedirector(wsm, agentIP, teamserverIP, domain, port, user, privateKeyPath, toastId, commands)
}

// RetrieveTeamserverPublicKey retrieves the public SSH key from the teamserver.
func RetrieveTeamserverPublicKey(teamserverIP, port, user, privateKeyPath string) (string, error) {
	sshClient, err := common.NewSSHClient(teamserverIP, "22", user, privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to establish SSH connection: %v", err)
	}
	// Execute a command to cat the public key file
	if _, err := sshClient.ExecuteCommand("ssh-keygen -t rsa -b 2048 -f ~/.ssh/id_rsa -N ''"); err != nil {
		return "", fmt.Errorf("failed to generate SSH key: %v", err)
	}

	publicKey, err := sshClient.ExecuteCommand("cat ~/.ssh/id_rsa.pub")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve public key: %v", err)
	}

	log.Info().Msgf("Retrieved public key from teamserver at %s", teamserverIP)
	return publicKey, nil
}

// DeploySSHTunnelService configures and starts an SSH reverse tunnel service on the teamserver.
func DeploySSHTunnelService(teamserverIP, redirectorIP, port, user, privateKeyPath, sshTunnelTemplatePath string) error {
	sshClient, err := common.NewSSHClient(teamserverIP, port, user, privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to establish SSH connection to teamserver: %v", err)
	}

	// Set redirector host to known hosts
	log.Info().Msg("Setting redirector host to known hosts")
	if _, err := sshClient.ExecuteCommand(fmt.Sprintf("ssh-keyscan -H %s >> ~/.ssh/known_hosts", redirectorIP)); err != nil {
		return fmt.Errorf("failed to set redirector host to known hosts: %v", err)
	}

	if err := GenerateAndTransferCert(sshClient, redirectorIP); err != nil {
		return fmt.Errorf("failed to generate and transfer certificate: %v", err)
	}

	// log.Info().Msg("SSH reverse tunnel service enabled and started on teamserver")
	log.Info().Msg("Certificate and key generated and transferred to redirector")

	if _, err := sshClient.ExecuteCommand("update-ca-certificates"); err != nil {
		return fmt.Errorf("failed to update CA certificates: %v", err)
	}

	return nil
}

// Function to execute OpenSSL command and transfer the certificate
func GenerateAndTransferCert(teamserverSSH *common.SSHClient, redirectorIP string) error {
	// Run the OpenSSL command on the teamserver to generate localhost.crt and localhost.key
	log.Info().Msg("openssl req -x509 -nodes -newkey rsa:2048 -keyout localhost.key -out localhost.crt -sha256 -days 365 -subj '/CN=localhost")
	genCertCmd := "openssl req -x509 -nodes -newkey rsa:2048 -keyout localhost.key -out localhost.crt -sha256 -days 365 -subj '/CN=localhost'"
	_, err := teamserverSSH.ExecuteCommand(genCertCmd)
	if err != nil {
		return fmt.Errorf("failed to generate certificate on teamserver: %v", err)
	}
	log.Info().Msg("Certificate and key generated on teamserver")

	// Transfer the certificate and key to the redirector
	log.Info().Msg("Transferring certificate from teamserver to redirector")
	scpCommand := fmt.Sprintf("scp /root/localhost.crt root@%s:/usr/local/share/ca-certificates/", redirectorIP)

	if _, err := teamserverSSH.ExecuteCommand(scpCommand); err != nil {
		return fmt.Errorf("failed to transfer certificate to redirector: %v", err)
	}

	return nil
}
