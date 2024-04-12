import React, { useState, useEffect, useCallback, useRef } from 'react';
import PhishingSetup from './PhishingSetup';
import { useAuth } from "../contexts/AuthContext";
import { Row, Col } from 'reactstrap';
import { ToastContainer, toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { useWebSocket } from '../contexts/WebSocketContext';

const PhishingInfrastructure = () => {
  const [agents, setAgents] = useState([]);
  const [domains, setDomains] = useState([]);

  useWebSocket();

  const { apiKey } = useAuth(); // Destructure apiKey from useAuth hook

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
  }, []);

  const handleSubmit = useCallback((payload) => {
    // Extract the values from the payload object  
    fetch("http://localhost:8080/phishing-setup", {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': apiKey,
      },
      body: JSON.stringify(payload),
    })
      .then(response => {
        if (!response.ok) {
          throw new Error("Phishing setup failed");
        }
      })
      .catch(error => {
        console.error("Phishing setup failed:", error);
      });
  }, [apiKey]);

  return (
    <Row>
      <Col lg="12">
        <ToastContainer />
        <PhishingSetup agents={agents} domains={domains} handleSubmit={handleSubmit} />
      </Col>
    </Row>
  );
};

export default PhishingInfrastructure;