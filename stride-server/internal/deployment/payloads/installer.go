package payloads

import (
	"fmt"

	"github.com/loosehose/stride/stride-server/internal/deployment/common"
	"github.com/loosehose/stride/stride-server/logging"
	"github.com/rs/zerolog/log"
)

func init() {
	logging.InitLogger()
}

func deployPayload(wsm *common.WebSocketManager, agentIP, user, privateKeyPath string, commands []string) error {
	sshClient, err := common.NewSSHClient(agentIP, "22", user, privateKeyPath)
	if err != nil {
		log.Error().Msgf("Failed to establish SSH connection to %s: %v", agentIP, err)
		common.NewLogMessage(common.Error, fmt.Sprintf("Error deploying payload: %v", err), "ERROR", wsm)
		return fmt.Errorf("error deploying payload: %v", err)
	}

	// Execute the commands over SSH
	totalCommands := len(commands)
	for i, cmd := range commands {
		// Notify the front end about the current step
		common.NewLogMessage(common.Exec, fmt.Sprintf("Progress: %d/%d", i+1, totalCommands), "EXEC", wsm)

		log.Info().Msgf("Executing: %s", cmd)
		output, err := sshClient.ExecuteCommand(cmd)
		if err != nil {
			log.Error().Msgf("Failed to execute command: %v", err)
			errorMessage := fmt.Sprintf("Failed to execute command: %v", err)
			log.Debug().Msgf("Output: %s", output)
			common.NewLogMessage(common.Error, errorMessage, "ERROR", wsm)
			return err
		}
	}

	successMessage := "Payload deployment completed successfully"
	common.NewLogMessage(common.Success, successMessage, "SUCCESS", wsm)
	return nil
}

func DeployShhhloader(wsm *common.WebSocketManager, agentIP, shellcodePath, user, privateKeyPath, commandLineOptions string) error {
	// Define the commands to clone Shhhloader and execute it
	cloneCmd := "git clone https://github.com/icyguider/Shhhloader.git"
	installDependenciesCmd := "cd Shhhloader/ && sudo apt install -y python3-pip && pip install -r requirements.txt"
	executeCmd := fmt.Sprintf("cd Shhhloader/ && python3 Shhhloader.py %s %s", shellcodePath, commandLineOptions)
	commands := []string{cloneCmd, installDependenciesCmd, executeCmd}

	return deployPayload(wsm, agentIP, user, privateKeyPath, commands)
}
