import React, { useState, useEffect, useMemo, useCallback } from 'react';
import TeamserverSetup from './TeamserverSetup';
import RedirectorSetup from './RedirectorSetup';
import PortForwardingSetup from './PortForwardingSetup';
import { useAuth } from "../contexts/AuthContext";
import {
  Row,
  Col,
} from 'reactstrap';
import { toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { useWebSocket } from '../contexts/WebSocketContext';

const Infrastructure = () => {
  const [agents, setAgents] = useState([]);
  const [domains, setDomains] = useState([]);
  const [selectedTeamserverAgent, setSelectedTeamserverAgent] = useState('');
  const [selectedTeamserverSoftware, setSelectedTeamserverSoftware] = useState([]);
  const [selectedRedirectorAgent, setSelectedRedirectorAgent] = useState('');
  const [selectedRedirectorSoftware, setSelectedRedirectorSoftware] = useState([]);
  const [selectedRedirectorDomain, setSelectedRedirectorDomain] = useState('');
  const [sourcePort, setSourcePort] = useState('');
  const [protocol, setProtocol] = useState('tcp');
  const [destinationPort, setDestinationPort] = useState('');
  const [sshKeys, setSshKeys] = useState([]);
  const [selectedRedirectorTeamserverAgent, setSelectedRedirectorTeamserverAgent] = useState('');

  useWebSocket();

  const { apiKey } = useAuth();

  useEffect(() => {
    fetch("http://localhost:8080/agents", {
      headers: {
        'X-API-Key': apiKey,
      },
    })
      .then(response => response.json())
      .then(data => setAgents(data.map(agent => ({ id: agent.id, name: agent.name, ip: agent.ip }))))
      .catch(error => console.error("Failed to fetch agents data:", error));

    fetch("http://localhost:8080/domains", {
      headers: {
        'X-API-Key': apiKey,
      },
    })
      .then(response => response.json())
      .then(setDomains)
      .catch(error => console.error("Failed to fetch domains data:", error));

    fetch("http://localhost:8080/ssh-keys", {
      headers: {
        'X-API-Key': apiKey,
      },
    })
      .then(response => response.json())
      .then(setSshKeys)
      .catch(error => console.error("Failed to fetch SSH keys:", error));
  }, []);

  const handleSubmitPortForwarding = useCallback(() => {
    const payload = {
      teamserverAgent: selectedTeamserverAgent,
      redirectorAgent: selectedRedirectorAgent,
      sourcePort,
      protocol,
      destinationPort,
    };

    fetch('http://localhost:8080/port-forwarding-setup', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': apiKey,
      },
      body: JSON.stringify(payload),
    })
      .then(response => {
        if (!response.ok) {
          throw new Error('Port forwarding setup failed');
        }
        return response.json();
      })
      .then(() => {
        toast.success('Port forwarding setup successfully!');
      })
      .catch(error => {
        console.error('Failed to submit port forwarding setup:', error);
        toast.error('Port forwarding setup failed!');
      });
  }, [selectedTeamserverAgent, selectedRedirectorAgent, sourcePort, protocol, destinationPort, apiKey]);

  const handleSubmit = useCallback((setupType) => {
    let url = '';
    let payload = {};

    if (setupType === 'Teamserver') {
      url = 'http://localhost:8080/teamserver-setup';
      payload = {
        agentIP: selectedTeamserverAgent,
        software: selectedTeamserverSoftware,
      };
    } else if (setupType === 'Redirector') {
      url = 'http://localhost:8080/redirector-setup';
      payload = {
        redirectorAgent: selectedRedirectorAgent,
        teamserverAgent: selectedRedirectorTeamserverAgent,
        software: selectedRedirectorSoftware,
        domain: selectedRedirectorDomain,
      };
    }

    fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': apiKey,
      },
      body: JSON.stringify(payload),
    })
      .then(response => {
        if (!response.ok) {
          throw new Error(`${setupType} setup failed`);
        }
      })
      .catch(error => {
        console.error(`${setupType} setup failed:`, error);
      });
  }, [selectedTeamserverAgent, selectedTeamserverSoftware, selectedRedirectorAgent, selectedRedirectorTeamserverAgent, selectedRedirectorSoftware, selectedRedirectorDomain, apiKey]);

  const teamserverAgentOptions = useMemo(() => agents.map(agent => ({ value: agent.ip, label: agent.name })), [agents]);
  const redirectorAgentOptions = useMemo(() => agents.map(agent => ({ value: agent.ip, label: agent.name })), [agents]);

  return (
    <Row>
      <Col lg="12">
        <TeamserverSetup
          agents={teamserverAgentOptions}
          selectedAgent={selectedTeamserverAgent}
          setSelectedAgent={setSelectedTeamserverAgent}
          selectedSoftware={selectedTeamserverSoftware}
          setSelectedSoftware={setSelectedTeamserverSoftware}
          handleSubmit={handleSubmit}
        />
        <RedirectorSetup
          agents={redirectorAgentOptions}
          domains={domains}
          selectedAgent={selectedRedirectorAgent}
          setSelectedAgent={setSelectedRedirectorAgent}
          selectedSoftware={selectedRedirectorSoftware}
          setSelectedSoftware={setSelectedRedirectorSoftware}
          selectedDomain={selectedRedirectorDomain}
          setSelectedDomain={setSelectedRedirectorDomain}
          selectedTeamserverAgent={selectedRedirectorTeamserverAgent}
          setSelectedTeamserverAgent={setSelectedRedirectorTeamserverAgent}
          handleSubmit={handleSubmit}
        />
        <PortForwardingSetup
          agents={agents}
          selectedTeamserverAgent={selectedTeamserverAgent}
          setSelectedTeamserverAgent={setSelectedTeamserverAgent}
          selectedRedirectorAgent={selectedRedirectorAgent}
          setSelectedRedirectorAgent={setSelectedRedirectorAgent}
        />
      </Col>
    </Row>
  );
};

export default Infrastructure;