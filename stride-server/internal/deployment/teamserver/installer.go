package teamserver

import (
	"fmt"
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

func deployTeamserver(wsm *common.WebSocketManager, agentIP, sshKeyNameOrID, user, privateKeyPath string, commands []string) error {
	sshClient, err := common.NewSSHClient(agentIP, "22", user, privateKeyPath)
	if err != nil {
		log.Error().Msgf("Failed to establish SSH connection to %s: %v", agentIP, err)
		common.NewLogMessage(common.Error, fmt.Sprintf("Error deploying teamserver: %v", err), "ERROR", wsm)
		return fmt.Errorf("error deploying teamserver: %v", err)
	}

	// Execute the commands over SSH
	totalCommands := len(commands)
	for i, cmd := range commands {
		// Notify the front end about the current step
		progressMessage := fmt.Sprintf("Progress: %d/%d", i+1, totalCommands)
		common.NewLogMessage(common.Exec, progressMessage, "EXEC", wsm)

		log.Info().Msgf("Executing: %s", cmd)
		output, err := sshClient.ExecuteCommand(cmd)
		if err != nil {
			log.Error().Msgf("Failed to execute command: %v", err)
			errorMessage := fmt.Sprintf("Failed to execute command: %v", err)
			common.NewLogMessage(common.Error, errorMessage, "ERROR", wsm)
			return err
		}

		if cmd == "cd ~/Mythic && echo $(sudo cat .env | grep MYTHIC_ADMIN_PASSWORD) | awk -F '\"' '{print $2}'" {
			// Send the password to the frontend user via WebSocket
			passwordMessage := fmt.Sprintf("Mythic admin password: %s", output)
			common.NewLogMessage(common.Info, "Check server logs for login information", "INFO", wsm)
			log.Info().Msg("----------------------------")
			log.Info().Msg("Mythic C2 Login Information")
			log.Info().Msg("----------------------------")
			log.Info().Msg("ssh -L 7443:localhost:7443 root@" + agentIP)
			log.Info().Msg("User: mythic_admin")
			log.Info().Msgf("Password: %s", passwordMessage)
			log.Info().Msg("----------------------------")
		}
	}

	successMessage := "Teamserver deployment completed successfully"
	log.Info().Msg(successMessage)
	common.NewLogMessage(common.Success, successMessage, "SUCCESS", wsm)
	return nil
}

func DeploySliverTeamserver(wsm *common.WebSocketManager, agentIP, sshKeyNameOrID, user, privateKeyPath string) error {
	// Get the absolute path of the installer.go file
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to determine the path of installer.go")
	}
	currentDir := filepath.Dir(currentFile)

	// Construct the absolute path to the sliver.service file
	sliverServiceFilePath := filepath.Join(currentDir, "sliver.service")

	sshClient, err := common.NewSSHClient(agentIP, "22", user, privateKeyPath)
	if err != nil {
		log.Error().Msgf("Failed to establish SSH connection to %s: %v", agentIP, err)
		common.NewLogMessage(common.Error, fmt.Sprintf("Error deploying Sliver: %v", err), "ERROR", wsm)
		return fmt.Errorf("error deploying Sliver: %v", err)
	}

	// Upload the sliver.service file to the remote agent
	if err := sshClient.TransferFile(sliverServiceFilePath, "/etc/systemd/system/sliver.service"); err != nil {
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to upload sliver.service file: %v", err), "ERROR", wsm)
		return err
	}

	// Commands to install Sliver on Ubuntu
	commands := []string{
		"sudo apt-get -y update",
		"sudo rm /var/lib/dpkg/lock",
		"sudo apt-get -y install build-essential mingw-w64 binutils-mingw-w64 g++-mingw-w64",
		"wget -O /usr/local/bin/sliver-server https://github.com/BishopFox/sliver/releases/download/v1.5.41/sliver-server_linux",
		"chmod 755 /usr/local/bin/sliver-server",
		"wget -O /usr/local/bin/sliver https://github.com/BishopFox/sliver/releases/download/v1.5.41/sliver-client_linux",
		"chmod 755 /usr/local/bin/sliver",
		"/usr/local/bin/sliver-server unpack --force",
		"mkdir -p /root/.sliver-client/configs",
		"sliver-server operator --name root --lhost localhost --save ~/.sliver-client/configs/",
		"systemctl daemon-reload && systemctl enable sliver.service && systemctl start sliver.service",
	}

	return deployTeamserver(wsm, agentIP, sshKeyNameOrID, user, privateKeyPath, commands)
}

func DeployMythicTeamserver(wsm *common.WebSocketManager, agentIP, sshKeyNameOrID, user, privateKeyPath string) error {
	commands := []string{
		"sudo apt-get -y update",
		"sudo apt-get -y install build-essential mingw-w64 binutils-mingw-w64 g++-mingw-w64 apt-transport-https ca-certificates curl gnupg-agent software-properties-common",
		"git clone https://github.com/its-a-feature/Mythic.git",
		"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -",
		"sudo add-apt-repository -y \"deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\"",
		"sudo apt-get -y update",
		"sudo apt-get -y install --no-install-recommends docker-ce docker-compose-plugin",
		"cd ~/Mythic && sudo make",
		"cd ~/Mythic && sudo ./mythic-cli install github https://github.com/MythicC2Profiles/http",
		"cd ~/Mythic && sudo ./mythic-cli install github https://github.com/MythicAgents/merlin",
		"cd ~/Mythic && sudo ./mythic-cli start",
		"cd ~/Mythic && echo $(sudo cat .env | grep MYTHIC_ADMIN_PASSWORD) | awk -F '\"' '{print $2}'",
	}

	return deployTeamserver(wsm, agentIP, sshKeyNameOrID, user, privateKeyPath, commands)
}

func DeployHavocC2Teamserver(wsm *common.WebSocketManager, agentIP, sshKeyNameOrID, user, privateKeyPath string) error {

	sshClient, err := common.NewSSHClient(agentIP, "22", user, privateKeyPath)
	if err != nil {
		log.Error().Msgf("Failed to establish SSH connection to %s: %v", agentIP, err)
		common.NewLogMessage(common.Error, fmt.Sprintf("Error deploying teamserver: %v", err), "ERROR", wsm)
		return fmt.Errorf("error deploying teamserver: %v", err)
	}

	// Get the latest Go version
	_, err = sshClient.ExecuteCommand("apt install -y jq")
	versionCmd := "curl -s https://go.dev/dl/?mode=json | jq -r '.[0].version'"
	version, err := sshClient.ExecuteCommand(versionCmd)
	if err != nil {
		log.Error().Msgf("Failed to get the latest Go version: %v", err)
		common.NewLogMessage(common.Error, fmt.Sprintf("Failed to get the latest Go version: %v", err), "ERROR", wsm)
	}
	version = strings.TrimSpace(version)

	installCommands := []string{
		"git clone https://github.com/HavocFramework/Havoc.git",
		"sudo add-apt-repository ppa:deadsnakes/ppa -y && sudo apt update -y && sudo apt install -y python3.10 python3.10-dev python3-pip git build-essential apt-utils cmake libfontconfig1 libglu1-mesa-dev libgtest-dev libspdlog-dev libboost-all-dev libncurses5-dev libgdbm-dev libssl-dev libreadline-dev libffi-dev libsqlite3-dev libbz2-dev mesa-common-dev qtbase5-dev qtchooser qt5-qmake qtbase5-dev-tools libqt5websockets5 libqt5websockets5-dev qtdeclarative5-dev qtbase5-dev libqt5websockets5-dev python3-dev libboost-all-dev mingw-w64 nasm",
		"cd /root/Havoc/teamserver && ./Install.sh",
		"add-apt-repository --remove golang-go",
		fmt.Sprintf("wget https://go.dev/dl/%s.linux-amd64.tar.gz && tar -C /usr/local -xzf %s.linux-amd64.tar.gz && ln -sf /usr/local/go/bin/go /usr/bin/go && rm %s.linux-amd64.tar.gz", version, version, version),
		"cd /root/Havoc/teamserver && go mod download golang.org/x/sys && go mod download github.com/ugorji/go && go mod tidy",
		"cd /root/Havoc/teamserver; GO111MODULE=\"on\" go build -ldflags=\"-s -w -X cmd.VersionCommit=$(git rev-parse HEAD)\" -o ../havoc main.go",
		"cd /root/Havoc && sudo setcap 'cap_net_bind_service=+ep' havoc",
		"git clone https://github.com/Ghost53574/havoc_profile_generator.git",
		"cd /root/havoc_profile_generator && pip3 install -r requirements.txt",
		"rm /root/havoc_profile_generator/profiles/bing_maps.json",
		"cd /root/havoc_profile_generator && python3 havoc_profile_generator.py -E -H 0.0.0.0 -S changeme.com -o /root/Havoc/profiles/random_profile.yaotl",
		"wget http://musl.cc/x86_64-w64-mingw32-cross.tgz && tar -xzf x86_64-w64-mingw32-cross.tgz -C /usr/bin && rm /root/x86_64-w64-mingw32-cross.tgz",
		"sed -i 's/user \".*\"/user \"admin\"/' /root/Havoc/profiles/random_profile.yaotl",
		"grep \"Password =\" /root/Havoc/profiles/random_profile.yaotl | awk -F '\"' '{print $2}' > /tmp/password.txt",
		"sed -i 's/Host = \"127.0.0.1\"/Host = \"0.0.0.0\"/' /root/Havoc/profiles/random_profile.yaotl",
		"sed -i 's#Compiler64 = \".*\"#Compiler64 = \"/usr/bin/x86_64-w64-mingw32-cross/bin/x86_64-w64-mingw32-gcc\"#' /root/Havoc/profiles/random_profile.yaotl",
		"sed -i 's#Compiler86 = \".*\"#Compiler86 = \"/usr/bin/x86_64-w64-mingw32-cross/bin/x86_64-w64-mingw32-gcc\"#' /root/Havoc/profiles/random_profile.yaotl",
		"grep \"Port =\" /root/Havoc/profiles/random_profile.yaotl | awk -F '\"' '{print $2}' > /tmp/port.txt",
		"tmux new-session -d -s havoc",
		"tmux send-keys -t havoc 'cd /root/Havoc && ./havoc server --profile ./profiles/random_profile.yaotl -v --debug' Enter",
	}

	var havocPassword string
	var havocPort string
	for i, command := range installCommands {
		progressMessage := fmt.Sprintf("Progress: %d/%d", i+1, len(installCommands))
		common.NewLogMessage(common.Exec, progressMessage, "EXEC", wsm)

		log.Info().Msgf("Executing: %s", command)
		_, err := sshClient.ExecuteCommand(command)
		if err != nil {
			log.Error().Msgf("Failed to execute command: %v", err)
			errorMessage := fmt.Sprintf("Failed to execute command: %v", err)
			common.NewLogMessage(common.Error, errorMessage, "ERROR", wsm)
			return err
		}

		if command == "grep \"Password =\" /root/Havoc/profiles/random_profile.yaotl | awk -F '\"' '{print $2}' > /tmp/password.txt" {
			// Read the password from the /tmp/password.txt file
			passwordOutput, err := sshClient.ExecuteCommand("cat /tmp/password.txt")
			if err != nil {
				log.Error().Msgf("Failed to read password file: %v", err)
				errorMessage := fmt.Sprintf("Failed to read password file: %v", err)
				common.NewLogMessage(common.Error, errorMessage, "ERROR", wsm)
				return err
			}

			havocPassword = strings.TrimSpace(passwordOutput)
		}

		if command == "grep \"Port =\" /root/Havoc/profiles/random_profile.yaotl | awk -F '\"' '{print $2}' > /tmp/port.txt" {
			// Read the port number from the /tmp/port.txt file
			portOutput, err := sshClient.ExecuteCommand("cat /tmp/port.txt")
			if err != nil {
				log.Error().Msgf("Failed to read port file: %v", err)
				errorMessage := fmt.Sprintf("Failed to read port file: %v", err)
				common.NewLogMessage(common.Error, errorMessage, "ERROR", wsm)
				return err
			}

			havocPort = strings.TrimSpace(portOutput)
		}
	}

	common.NewLogMessage(common.Info, "Check server logs for login information", "INFO", wsm)
	log.Info().Msg("----------------------------")
	log.Info().Msg("Havoc C2 Login Information")
	log.Info().Msg("----------------------------")
	log.Info().Msgf("Host: %s", agentIP)
	log.Info().Msgf("Port: %s", havocPort)
	log.Info().Msg("User: admin")
	log.Info().Msgf("Password: %s", havocPassword)
	log.Info().Msg("----------------------------")

	successMessage := "Havoc C2 teamserver deployment completed successfully"
	log.Info().Msg(successMessage)
	common.NewLogMessage(common.Success, successMessage, "SUCCESS", wsm)
	return nil
}
