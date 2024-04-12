import React, { useEffect, useState, useCallback } from "react";
import Select from 'react-select';
import { useAuth } from "../contexts/AuthContext";
import {
    Card,
    CardHeader,
    CardBody,
    CardTitle,
    Table,
    Button,
    Form,
    FormGroup,
    Input,
    Row,
    Col,
    Label,
    Collapse,
} from "reactstrap";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTrash, faChevronDown, faChevronUp } from '@fortawesome/free-solid-svg-icons';
import { customStyles } from './selectedStyles';
import { useInputStyles } from './inputStyles';


const Agents = () => {
    const [agents, setAgents] = useState([]);
    const [sshKeys, setSshKeys] = useState([]);
    const [showForm, setShowForm] = useState(false);
    const [newAgentName, setNewAgentName] = useState("");
    const [newAgentSSHKey, setNewAgentSSHKey] = useState("");
    const [selectedSSHKeys, setSelectedSSHKeys] = useState([]);
    const [selectedSize, setSelectedSize] = useState("s-1vcpu-1gb");

    const { apiKey } = useAuth(); // Destructure apiKey from useAuth hook

    const agentNameStyles = useInputStyles();

    const fetchAgents = useCallback(() => {
        fetch("http://localhost:8080/agents", {
            headers: {
                'X-API-Key': apiKey,
            },
        })
            .then(response => response.json())
            .then(data => setAgents(data))
            .catch(error => console.error("Failed to fetch agents data:", error));
    }, [apiKey]);

    useEffect(() => {
        fetchAgents();

        fetch("http://localhost:8080/ssh-keys", {
            headers: {
                'X-API-Key': apiKey,
            },
        })
            .then(response => response.json())
            .then(data => setSshKeys(data))
            .catch(error => console.error("Failed to fetch SSH keys:", error));
    }, [fetchAgents, apiKey]);

    const sizeOptions = [
        { value: "s-1vcpu-1gb", label: "1 vCPU, 1GB RAM" },
        { value: "s-1vcpu-2gb", label: "1 vCPU, 2GB RAM" },
        { value: "s-2vcpu-2gb", label: "2 vCPUs, 2GB RAM" },
        { value: "s-2vcpu-4gb", label: "2 vCPUs, 4GB RAM" },
    ];

    const handleNameChange = e => setNewAgentName(e.target.value);
    const handleSSHKeyChange = selectedOptions => {
        setSelectedSSHKeys(selectedOptions);
    };

    const handleSubmit = useCallback((e) => {
        e.preventDefault();

        fetch("http://localhost:8080/agents", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                'X-API-Key': apiKey
            },
            body: JSON.stringify({
                agent: newAgentName,
                size: selectedSize,
                ssh_keys: selectedSSHKeys.map(option => option.value),
            }),
        })
        .then((response) => {
            if (response.ok) {
                // Start polling for agents data
                const pollAgents = setInterval(() => {
                    fetch("http://localhost:8080/agents", {
                        headers: {
                            'X-API-Key': apiKey,
                        },
                    })
                        .then(response => response.json())
                        .then(data => {
                            setAgents(data);
                            // Check if the newly created agent exists in the fetched data
                            const newAgent = data.find(agent => agent.name === newAgentName);
                            if (newAgent) {
                                // Clear the polling interval when the new agent is found
                                clearInterval(pollAgents);
                            }
                        })
                        .catch(error => console.error("Failed to fetch agents data:", error));
                }, 5000); // Poll every 5 seconds

                setNewAgentName("");
                setNewAgentSSHKey("");
                setShowForm(false);
            } else {
                throw new Error('Failed to create the agent.');
            }
        })
        .catch((error) => {
            console.error("Failed to create agent:", error);
        });
}, [apiKey, newAgentName, selectedSSHKeys, selectedSize]);

    const handleDelete = useCallback((dropletID) => {
        fetch(`http://localhost:8080/agents/${dropletID}`, {
            method: "DELETE",
            headers: {
                "Content-Type": "application/json",
                'X-API-Key': apiKey
            },
        })
            .then((response) => {
                if (response.ok) {
                    setAgents(prevAgents => prevAgents.filter(agent => agent.id !== dropletID));
                } else {
                    throw new Error('Failed to delete the agent.');
                }
            })
            .catch(error => {
                console.error("Failed to delete agent:", error);
            });
    }, [apiKey]);

    const toggleFormVisibility = () => setShowForm(!showForm);

    return (
        <Card>
            <CardHeader>
                <Row>
                    <Col xs="12" sm="6">
                        <CardTitle tag="h4">Agents</CardTitle>
                    </Col>
                    <Col xs="12" sm="6" className="text-right">
                        <Button color="info" size="sm" onClick={toggleFormVisibility}>
                            {showForm ? (
                                <FontAwesomeIcon icon={faChevronUp} />
                            ) : (
                                <FontAwesomeIcon icon={faChevronDown} />
                            )}
                        </Button>
                    </Col>
                </Row>
            </CardHeader>
            <Collapse isOpen={showForm}>
                <CardBody>
                    <Form onSubmit={handleSubmit} className="mb-2">
                        <FormGroup>
                            <Label for="sshKeySelect">Agent Name</Label>
                            <Input
                                type="text"
                                name="name"
                                placeholder="foobarTeamserver"
                                value={newAgentName}
                                style={agentNameStyles.getInputStyles()}
                                onChange={handleNameChange}
                                onFocus={agentNameStyles.handleFocus}
                                onBlur={agentNameStyles.handleBlur}
                            />
                        </FormGroup>
                        <FormGroup>
                            <Label for="sshKeySelect">Select SSH Key(s)</Label>
                            <Select
                                isMulti
                                name="sshKeys"
                                options={sshKeys.map(key => ({ value: key.id, label: key.name }))}
                                className="basic-multi-select"
                                classNamePrefix="select"
                                styles={customStyles}
                                onChange={handleSSHKeyChange}
                                value={selectedSSHKeys}
                            />
                        </FormGroup>
                        <FormGroup>
                            <Label for="sizeSelect">Select Size</Label>
                            <Select
                                name="size"
                                options={sizeOptions}
                                className="basic-select"
                                classNamePrefix="select"
                                styles={customStyles}
                                onChange={selectedOption => setSelectedSize(selectedOption.value)}
                                value={sizeOptions.find(option => option.value === selectedSize)}
                            />
                        </FormGroup>
                        <Row className="justify-content-end mt-3">
                            <Col sm="auto">
                                <Button color="info" size="sm" type="submit">Submit</Button>
                            </Col>
                        </Row>
                    </Form>
                </CardBody>
            </Collapse>
            <CardBody>
                <Table className="tablesorter" responsive>
                    <thead className="text-primary">
                        <tr>
                            <th>Name</th>
                            <th>IP Address</th>
                            <th>Created</th>
                            <th>Tags</th>
                            <th>Action</th>
                        </tr>
                    </thead>
                    <tbody>
                        {(agents || []).map((agent, index) => (
                            <tr key={index}>
                                <td>{agent.name}</td>
                                <td>{agent.ip}</td>
                                <td>{new Date(agent.created).toLocaleDateString()}</td>
                                <td>{agent.tags.join(", ")}</td>
                                <td>
                                    <FontAwesomeIcon
                                        icon={faTrash}
                                        style={{ cursor: 'pointer', color: 'lightgrey' }}
                                        onClick={() => handleDelete(agent.id)}
                                        onMouseOver={(e) => e.target.style.color = 'red'}
                                        onMouseOut={(e) => e.target.style.color = 'lightgrey'}
                                    />
                                </td>
                            </tr>
                        ))}
                        {(agents || []).length === 0 && (
                            <tr>
                                <td colSpan="5">No agents found.</td>
                            </tr>
                        )}
                    </tbody>
                </Table>
            </CardBody>
        </Card>
    );
};

export default Agents;