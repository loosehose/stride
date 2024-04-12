import React, { useState, useCallback } from 'react';
import Select from 'react-select';
import { useAuth } from "../contexts/AuthContext";
import { Card, CardHeader, CardBody, CardTitle, FormGroup, Label, Button, Input, Col, Row } from 'reactstrap';
import 'react-toastify/dist/ReactToastify.css';
import { customStyles } from './selectedStyles';
import { useInputStyles } from './inputStyles';

const PortForwardingSetup = ({ agents, selectedTeamserverAgent, setSelectedTeamserverAgent, selectedRedirectorAgent, setSelectedRedirectorAgent }) => {
    const [sourcePort, setSourcePort] = useState('');
    const [protocol, setProtocol] = useState('tcp');
    const [destinationPort, setDestinationPort] = useState('');
    const { apiKey } = useAuth();
    const sourcePortInputStyles = useInputStyles();
    const destinationPortInputStyles = useInputStyles();

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
        .catch(error => {
            console.error('Failed to submit port forwarding setup:', error);
        });
    }, [selectedTeamserverAgent, selectedRedirectorAgent, sourcePort, protocol, destinationPort, apiKey]);

    const protocolOptions = [
        { value: 'tcp', label: 'TCP' },
        { value: 'udp', label: 'UDP' },
    ];

    return (
        <Card>
            <CardHeader><CardTitle tag="h4">Port Forwarding Setup</CardTitle></CardHeader>
            <CardBody>
                <FormGroup>
                    <Label>Teamserver Agent</Label>
                    <Select
                        value={agents.find(agent => agent.ip === selectedTeamserverAgent)}
                        onChange={selectedOption => setSelectedTeamserverAgent(selectedOption.ip)}
                        options={agents.map(agent => ({ value: agent.ip, label: agent.name }))}
                        styles={customStyles}
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Redirector Agent</Label>
                    <Select
                        value={agents.find(agent => agent.ip === selectedRedirectorAgent)}
                        onChange={selectedOption => setSelectedRedirectorAgent(selectedOption.ip)}
                        options={agents.map(agent => ({ value: agent.ip, label: agent.name }))}
                        styles={customStyles}
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Source Port</Label>
                    <Input
                        type="text"
                        value={sourcePort}
                        onChange={e => setSourcePort(e.target.value)}
                        style={sourcePortInputStyles.getInputStyles()}
                        onFocus={sourcePortInputStyles.handleFocus}
                        onBlur={sourcePortInputStyles.handleBlur}
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Protocol</Label>
                    <Select
                        value={protocolOptions.find(option => option.value === protocol)}
                        onChange={selectedOption => setProtocol(selectedOption.value)}
                        options={protocolOptions}
                        styles={customStyles}
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Destination Port</Label>
                    <Input
                        type="text"
                        value={destinationPort}
                        onChange={e => setDestinationPort(e.target.value)}
                        style={destinationPortInputStyles.getInputStyles()}
                        onFocus={destinationPortInputStyles.handleFocus}
                        onBlur={destinationPortInputStyles.handleBlur}
                    />
                </FormGroup>
                <Row className="justify-content-end mt-3">
                <Col sm="auto">
                    <Button color="info" size="sm" onClick={handleSubmitPortForwarding}>Submit</Button>
                </Col>
                </Row>
            </CardBody>
        </Card>
    );
};

export default PortForwardingSetup;