package portforward

import (
	"fmt"

	"github.com/loosehose/stride/stride-server/internal/deployment/common"
	"github.com/loosehose/stride/stride-server/logging"
	"github.com/rs/zerolog/log"
)

func init() {
	logging.InitLogger()
}

// DeployPortForwarding sets up port forwarding using iptables.
func DeployPortForwarding(wsm *common.WebSocketManager, redirectorAgentIP, teamserverAgentIP, sourcePort, protocol, destinationPort, privateKeyPath string) error {
	// Commands to modify sysctl.conf, apply changes, and setup iptables rules for port forwarding
	commands := []string{
		fmt.Sprintf("echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf"),
		"sysctl -p",
		fmt.Sprintf("iptables -t nat -A PREROUTING -p %s --dport %s -d %s -j DNAT --to-destination %s:%s", protocol, sourcePort, redirectorAgentIP, teamserverAgentIP, destinationPort),
		fmt.Sprintf("iptables -t nat -A POSTROUTING -p %s --dport %s -j SNAT --to-source %s", protocol, sourcePort, redirectorAgentIP),
	}

	// Execute the commands on the redirector agent
	sshClient, err := common.NewSSHClient(redirectorAgentIP, "22", "root", privateKeyPath)
	if err != nil {
		log.Fatal().Msgf("Failed to establish SSH connection to %s: %v", redirectorAgentIP, err)
		common.NewLogMessage(common.Error, fmt.Sprintf("Error setting up port forwarding: %v", err), "ERROR", wsm)
		return fmt.Errorf("error setting up port forwarding: %v", err)
	}

	for _, cmd := range commands {
		log.Info().Msgf("Executing: %s", cmd)
		common.NewLogMessage(common.Exec, cmd, "EXEC", nil)
		if _, err := sshClient.ExecuteCommand(cmd); err != nil {
			return fmt.Errorf("failed to execute command '%s': %v", cmd, err)
		}
	}

	return nil
}
