import React, { useState, useEffect } from 'react';
import PayloadSetup from './PayloadSetup';
import { useAuth } from "../contexts/AuthContext";
import { Row, Col } from 'reactstrap';
import 'react-toastify/dist/ReactToastify.css';
import { useWebSocket } from '../contexts/WebSocketContext';

const PayloadInfrastructure = () => {
  const [agents, setAgents] = useState([]); 

  useWebSocket();

  const { apiKey } = useAuth(); // Use apiKey from AuthContext for API calls

  useEffect(() => {
    fetch("http://localhost:8080/agents", {
      headers: {
        'X-API-Key': apiKey,
      },
    })
      .then(response => response.json())
      .then(data => setAgents(data.map(agent => ({ id: agent.id, name: agent.name, ip: agent.ip }))))
      .catch(error => console.error("Failed to fetch agents data:", error));
  }, []);

  const handleSubmit = async (payload) => {

    try {
      const response = await fetch(`http://localhost:8080/payload-setup`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Key': apiKey,
          },
        body: JSON.stringify(payload),
      });
    } catch (error) {
      console.error("Payload setup failed:", error);
    }
  };

  return (
    <div className="content">
      <Row>
        <Col md="12">
          <PayloadSetup agents={agents} handleSubmit={handleSubmit} />
        </Col>
      </Row>
    </div>
  );
};

export default PayloadInfrastructure;
