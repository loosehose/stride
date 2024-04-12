package phishing

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/loosehose/stride/stride-server/internal/deployment/common"
	"github.com/loosehose/stride/stride-server/logging"
	"github.com/rs/zerolog/log"
)

func init() {
	logging.InitLogger()
}

const (
	sshPort           = "22"
	gophishOutputFile = "/tmp/gophish_output.txt"
	evilGophishRepo   = "https://github.com/fin3ss3g0d/evilgophish.git"
)

type DeploymentConfig struct {
	APIToken       string
	AgentIP        string
	User           string
	PrivateKeyPath string
	RootDomain     string
	Subdomains     []string
	RootDomainBool string
	RedirectURL    string
	FeedBool       string
	RIDReplacement string
	BlacklistBool  string
	CertsPath      string
}

// DeployEvilGoPhish configures and deploys EvilGoPhish on a remote agent.
func DeployEvilGoPhish(wsm *common.WebSocketManager, config DeploymentConfig) (gophishPassword, txtRecord string, err error) {
	sshClient, err := common.NewSSHClient(config.AgentIP, sshPort, config.User, config.PrivateKeyPath)
	if err != nil {
		log.Printf("Failed to establish SSH connection to %s: %v", config.AgentIP, err)
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to establish SSH connection to %s: %v", config.AgentIP, err), "ERROR", wsm)
		return "", "", fmt.Errorf("error deploying EvilGoPhish: %v", err)
	}

	deploymentSteps := []func(*common.SSHClient, *common.WebSocketManager, DeploymentConfig) error{
		processAndTransferAuthHook,
		cloneEvilGophishRepo,
		transferSetupScript,
		installDependencies,
		generateCerts,
		setupApache,
		setupGophish,
		setupEvilginx3,
	}

	totalSteps := len(deploymentSteps)
	for i, step := range deploymentSteps {
		common.NewLogMessage(common.Exec, fmt.Sprintf("Progress: %d/%d", i+1, totalSteps), "EXEC", wsm)
		if err := step(sshClient, wsm, config); err != nil {
			return "", "", err
		}
	}

	// Start Gophish and get the password
	gophishPassword, err = startEvilGophishAndGetPassword(sshClient, wsm, config)
	if err != nil {
		return "", "", err
	}

	common.NewLogMessage(common.Info, fmt.Sprintf("Gophish admin password: %s", gophishPassword), "INFO", wsm)
	log.Printf("Gophish admin password: %s", gophishPassword)
	log.Info().Msg("----------------------------")
	log.Info().Msg("GoPhish Login Information")
	log.Info().Msg("----------------------------")
	log.Info().Msg("ssh -L 3333:localhost:333 root@" + config.AgentIP)
	log.Info().Msg("User: admin")
	log.Info().Msgf("Password: %s", gophishPassword)
	log.Info().Msg("----------------------------")
	common.NewLogMessage(common.Info, "EvilGoPhish setup completed successfully", "SUCCESS", wsm)
	return gophishPassword, txtRecord, nil
}

func processAndTransferAuthHook(sshClient *common.SSHClient, wsm *common.WebSocketManager, config DeploymentConfig) error {
	currentDir := getCurrentDir()
	authHookTemplatePath := filepath.Join(currentDir, "auth-hook.sh.tmpl")

	authHookTemplate, err := common.LoadFile(authHookTemplatePath)
	if err != nil {
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to load auth-hook.sh.tmpl: %v", err), "ERROR", wsm)
		return fmt.Errorf("failed to load auth-hook.sh.tmpl: %v", err)
	}

	authHookContent := strings.Replace(authHookTemplate, "{{DO_API_TOKEN}}", config.APIToken, -1)

	tempFile, err := createTempFile("auth-hook.sh", authHookContent)
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	if err := sshClient.TransferFile(tempFile.Name(), "/tmp/auth-hook.sh"); err != nil {
		return fmt.Errorf("failed to transfer auth-hook.sh: %v", err)
	}

	cleanupHookTemplatePath := filepath.Join(currentDir, "cleanup-hook.sh.tmpl")
	if err := sshClient.TransferFile(cleanupHookTemplatePath, "/tmp/cleanup-hook.sh"); err != nil {
		return fmt.Errorf("failed to transfer cleanup-hook.sh: %v", err)
	}

	return nil
}

func cloneEvilGophishRepo(sshClient *common.SSHClient, wsm *common.WebSocketManager, config DeploymentConfig) error {
	_, err := executeCommand(sshClient, wsm, fmt.Sprintf("git clone %s", evilGophishRepo))
	if err != nil {
		return err
	}
	return nil
}

func transferSetupScript(sshClient *common.SSHClient, wsm *common.WebSocketManager, config DeploymentConfig) error {
	currentDir := getCurrentDir()
	setupPath := filepath.Join(currentDir, "setup.sh")
	if err := sshClient.TransferFile(setupPath, "/root/evilgophish/setup.sh"); err != nil {
		log.Printf("Failed to transfer setup.sh file: %v", err)
		return err
	}
	return nil
}

func installDependencies(sshClient *common.SSHClient, wsm *common.WebSocketManager, config DeploymentConfig) error {
	installCommands := []string{
		"sudo rm /var/lib/apt/lists/lock",
		"apt-get update -y && sleep 5",
		"sudo rm /var/lib/dpkg/lock-frontend",
		"apt-get install apache2 build-essential letsencrypt certbot wget git net-tools tmux openssl jq -y",
	}

	for _, cmd := range installCommands {
		_, err := executeCommand(sshClient, wsm, cmd)
		if err != nil {
			log.Printf("Command execution failed: %v", err)
			// Handle the error or log a warning, depending on your requirements
		}
	}

	// Get the latest Go version
	versionCmd := "curl -s https://go.dev/dl/?mode=json | jq -r '.[0].version'"
	version, err := executeCommand(sshClient, wsm, versionCmd)
	if err != nil {
		log.Printf("Failed to get the latest Go version: %v", err)
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to get the latest Go version: %v", err), "ERROR", wsm)
	}
	version = strings.TrimSpace(version)

	goInstallCommands := []string{
		fmt.Sprintf("wget https://go.dev/dl/%s.linux-amd64.tar.gz", version),
		fmt.Sprintf("tar -C /usr/local -xzf %s.linux-amd64.tar.gz", version),
		"ln -sf /usr/local/go/bin/go /usr/bin/go",
		fmt.Sprintf("rm %s.linux-amd64.tar.gz", version),
	}

	for _, cmd := range goInstallCommands {
		_, err := executeCommand(sshClient, wsm, cmd)
		if err != nil {
			log.Printf("Go installation command failed: %v", err)
			common.NewLogMessage(common.Error, fmt.Sprintf("Go installation command failed: %v", err), "ERROR", wsm)
		}
	}

	return nil
}

func generateCerts(sshClient *common.SSHClient, wsm *common.WebSocketManager, config DeploymentConfig) error {
	_, err := executeCommand(sshClient, wsm, "chmod +x /tmp/auth-hook.sh && chmod +x /tmp/cleanup-hook.sh")
	if err != nil {
		log.Printf("Failed to set execute permissions on hook scripts: %v", err)
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to set execute permissions on hook scripts: %v", err), "ERROR", wsm)
	}

	domainArgs := fmt.Sprintf("-d %s", config.RootDomain)
	for _, subdomain := range config.Subdomains {
		fullDomain := fmt.Sprintf("%s.%s", subdomain, config.RootDomain)
		domainArgs = fmt.Sprintf("%s -d %s", domainArgs, fullDomain)
	}

	certbotCmd := fmt.Sprintf("certbot certonly --manual --preferred-challenges=dns --manual-auth-hook /tmp/auth-hook.sh --manual-cleanup-hook /tmp/cleanup-hook.sh --email admin@%s --server https://acme-v02.api.letsencrypt.org/directory --agree-tos %s --no-eff-email --manual-public-ip-logging-ok", config.RootDomain, domainArgs)
	_, _ = executeCommand(sshClient, wsm, certbotCmd)

	return nil
}

func setupApache(sshClient *common.SSHClient, wsm *common.WebSocketManager, config DeploymentConfig) error {
	apacheCommands := []string{
		"a2enmod proxy",
		"a2enmod proxy_http",
		"a2enmod proxy_balancer",
		"a2enmod lbmethod_byrequests",
		"a2enmod rewrite",
		"a2enmod ssl",
		fmt.Sprintf("sed 's/ServerAlias evilginx3.template/ServerAlias %s/g' /root/evilgophish/conf/000-default.conf.template > /root/evilgophish/000-default.conf", strings.Join(config.Subdomains, " ")),
		fmt.Sprintf("sed -i 's|SSLCertificateFile|SSLCertificateFile /etc/letsencrypt/live/%s/fullchain.pem|g' /root/evilgophish/000-default.conf", config.RootDomain),
		fmt.Sprintf("sed -i 's|SSLCertificateKeyFile|SSLCertificateKeyFile /etc/letsencrypt/live/%s/privkey.pem|g' /root/evilgophish/000-default.conf", config.RootDomain),
		"sed -i 's|Listen 80||g' /etc/apache2/ports.conf",
		fmt.Sprintf("sed 's|https://en.wikipedia.org/|%s|g' /root/evilgophish/conf/redirect.rules.template > /root/evilgophish/redirect.rules", config.RedirectURL),
		"cp /root/evilgophish/000-default.conf /etc/apache2/sites-enabled/",
		"cp /root/evilgophish/conf/blacklist.conf /etc/apache2/",
		"cp /root/evilgophish/redirect.rules /etc/apache2/",
		"rm /root/evilgophish/redirect.rules /root/evilgophish/000-default.conf",
	}

	for _, cmd := range apacheCommands {
		_, err := executeCommand(sshClient, wsm, cmd)
		if err != nil {
			log.Printf("Apache setup command failed: %v", err)
			common.NewLogMessage(common.Error, fmt.Sprintf("Apache setup command failed: %v", err), "ERROR", wsm)
		}
	}

	return nil
}

func setupGophish(sshClient *common.SSHClient, wsm *common.WebSocketManager, config DeploymentConfig) error {
	gophishCommands := []string{
		"cp /etc/hosts /etc/hosts.bak",
		fmt.Sprintf("sed -i 's|127.0.0.1.*|127.0.0.1 localhost %s %s|g' /etc/hosts", strings.Join(config.Subdomains, " "), config.RootDomain),
		"cp /etc/resolv.conf /etc/resolv.conf.bak",
		"rm /etc/resolv.conf",
		"ln -sf /run/systemd/resolve/resolv.conf /etc/resolv.conf",
		"systemctl stop systemd-resolved",
	}

	for _, cmd := range gophishCommands {
		_, err := executeCommand(sshClient, wsm, cmd)
		if err != nil {
			log.Printf("Gophish setup command failed: %v", err)
			common.NewLogMessage(common.Error, fmt.Sprintf("Gophish setup command failed: %v", err), "ERROR", wsm)
		}
	}

	if config.FeedBool == "true" {
		_, err := executeCommand(sshClient, wsm, "sed -i 's|\"feed_enabled\": false,|\"feed_enabled\": true,|g' /root/evilgophish/gophish/config.json")
		if err != nil {
			log.Printf("Failed to enable feed in Gophish config: %v", err)
			// Handle the error or log a warning, depending on your requirements
		}
		_, err = executeCommand(sshClient, wsm, "cd /root/evilgophish/evilfeed && go build")
		if err != nil {
			common.NewLogMessage(common.Error, fmt.Sprintf("Failed to build EvilFeed: %v", err), "ERROR", wsm)
		}
	}

	_, err := executeCommand(sshClient, wsm, fmt.Sprintf("find /root/evilgophish -type f -exec sed -i 's|client_id|%s|g' {} \\;", config.RIDReplacement))
	if err != nil {
		log.Printf("Failed to replace client_id: %v", err)
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to replace client_id: %v", err), "ERROR", wsm)
	}

	_, err = executeCommand(sshClient, wsm, "cd /root/evilgophish/gophish && go build")
	if err != nil {
		log.Printf("Failed to build Gophish: %v", err)
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to build Gophish: %v", err), "ERROR", wsm)
	}

	return nil
}

func setupEvilginx3(sshClient *common.SSHClient, wsm *common.WebSocketManager, config DeploymentConfig) error {
	evilginx3Commands := []string{
		"cd /root/evilgophish/evilginx3 && go build -o evilginx3",
	}

	for _, cmd := range evilginx3Commands {
		_, err := executeCommand(sshClient, wsm, cmd)
		if err != nil {
			log.Printf("Evilginx3 setup command failed: %v", err)
			common.NewLogMessage(common.Error, fmt.Sprintf("Evilginx3 setup command failed: %v", err), "ERROR", wsm)
		}
	}

	return nil
}

func startEvilGophishAndGetPassword(sshClient *common.SSHClient, wsm *common.WebSocketManager, config DeploymentConfig) (string, error) {
	if _, err := executeCommand(sshClient, wsm, "tmux new-session -d -s gophish"); err != nil {
		return "", err
	}

	if _, err := executeCommand(sshClient, wsm, "tmux new-session -d -s evilginx"); err != nil {
		return "", err
	}

	if _, err := executeCommand(sshClient, wsm, fmt.Sprintf("tmux send-keys -t gophish 'cd /root/evilgophish/gophish/ && ./gophish > %s 2>&1' Enter", gophishOutputFile)); err != nil {
		return "", err
	}

	time.Sleep(10 * time.Second)

	password, err := executeCommand(sshClient, wsm, fmt.Sprintf("grep 'Please login with the username admin and the password' %s | awk '{print substr($NF, 1, length($NF)-1)}'", gophishOutputFile))
	if err != nil {
		return "", fmt.Errorf("failed to extract Gophish password: %v", err)
	}

	if _, err := executeCommand(sshClient, wsm, "tmux send-keys -t evilginx 'cd /root/evilgophish/evilginx3 && ./evilginx3 -g /root/evilgophish/gophish/gophish.db -p ./legacy_phishlets' Enter"); err != nil {
		return "", err
	}

	return strings.TrimSpace(password), nil
}

func executeCommand(sshClient *common.SSHClient, wsm *common.WebSocketManager, command string) (string, error) {
	log.Printf("Executing command: %s", command)
	output, err := sshClient.ExecuteCommand(command)
	if err != nil {
		log.Printf("Command execution failed: %v, output: %s", err, output)
		return output, err
	}
	log.Printf("Command executed successfully: %s", output)
	return output, nil
}

func getCurrentDir() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Dir(currentFile)
}

func createTempFile(prefix, content string) (*os.File, error) {
	tempFile, err := ioutil.TempFile(os.TempDir(), prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}

	if _, err = tempFile.Write([]byte(content)); err != nil {
		return nil, fmt.Errorf("failed to write to temporary file: %v", err)
	}

	return tempFile, nil
}
