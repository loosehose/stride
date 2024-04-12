import React, { useEffect, useState } from 'react';
import Select from 'react-select';
import { useAuth } from '../contexts/AuthContext';
import {
    Card,
    CardHeader,
    CardBody,
    CardTitle,
    Table,
    Button,
    FormGroup,
    Label,
    Input,
    Form,
    Collapse,
    Row,
    Col,
} from 'reactstrap';
import { toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTrash, faChevronDown, faChevronUp } from '@fortawesome/free-solid-svg-icons';
import { customStyles } from './selectedStyles';
import { useInputStyles } from './inputStyles';

const Domains = () => {
    const [domains, setDomains] = useState([]);
    const [agents, setAgents] = useState([]);
    const [expandedDomain, setExpandedDomain] = useState(null);
    const [subdomainName, setSubdomainName] = useState('');
    const [selectedAgent, setSelectedAgent] = useState(null);

    const domainNameStyles = useInputStyles();

    const { apiKey } = useAuth();

    useEffect(() => {
        fetchDomains();
        fetchAgents();
    }, []);

    const fetchDomains = () => {
        fetch(`http://localhost:8080/domains`, {
            headers: {
                'X-API-Key': apiKey,
            },
        })
            .then((response) => response.json())
            .then((data) => setDomains(data))
            .catch((error) => console.error('Failed to fetch domains:', error));
    };

    const fetchAgents = () => {
        fetch(`http://localhost:8080/agents`, {
            headers: {
                'X-API-Key': apiKey,
            },
        })
            .then((response) => response.json())
            .then((data) => setAgents(data))
            .catch((error) => console.error('Failed to fetch agents:', error));
    };

    const handleCreateSubdomain = (domainName) => {
        const agent = agents.find((a) => a.name === selectedAgent?.value);
        if (!agent) {
            toast.error('Please select a valid agent.');
            return;
        }

        const payload = {
            domainName: domainName,
            subdomainName: subdomainName,
            recordType: 'A',
            data: agent.ip,
        };

        fetch(`http://localhost:8080/subdomains`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': apiKey,
            },
            body: JSON.stringify(payload),
        })
            .then((response) => {
                if (response.status === 201 || response.ok) {
                    return response.json();
                } else {
                    throw new Error('Failed to create the subdomain');
                }
            })
            .then((data) => {
                const updatedDomains = domains.map(domain => {
                    if (domain.name === domainName) {
                        return {
                            ...domain,
                            records: [...domain.records, data]
                        };
                    }
                    return domain;
                });
                setDomains(updatedDomains);
                setSubdomainName('');
                setSelectedAgent(null);
            })
            .catch((error) => {
                console.error('Failed to create subdomain:', error);
            });
    };

    const handleDeleteSubdomain = (domainName, subdomainName) => {
        const subdomain = subdomainName.split('.')[0];

        fetch(`http://localhost:8080/subdomains/${domainName}/${subdomain}`, {
            method: 'DELETE',
            headers: {
                'X-API-Key': apiKey,
            },
        })
            .then((response) => {
                if (response.ok) {
                    const updatedDomains = domains.map(domain => {
                        if (domain.name === domainName) {
                            return {
                                ...domain,
                                records: domain.records.filter(record => record.name !== subdomainName)
                            };
                        }
                        return domain;
                    });
                    setDomains(updatedDomains);
                } else {
                    throw new Error('Failed to delete the subdomain');
                }
            })
            .catch((error) => {
                console.error('Failed to delete subdomain:', error);
            });
    };

    const toggleExpand = (domain) => {
        setExpandedDomain(expandedDomain === domain ? null : domain);
        setSubdomainName('');
        setSelectedAgent(null);
    };

    return (
        <Card>
            <CardHeader>
                <CardTitle tag="h4">Domains</CardTitle>
            </CardHeader>
            <CardBody>
                <Table className="tablesorter" responsive>
                    <thead className="text-primary">
                        <tr>
                            <th>Domain Name</th>
                            <th>Subdomains</th>
                            <th>Action</th>
                        </tr>
                    </thead>
                    <tbody>
                        {domains.map((domain) => (
                            <React.Fragment key={domain.name}>
                                <tr>
                                    <td>{domain.name}</td>
                                    <td>{domain.records.length}</td>
                                    <td>
                                        <Button color="info" size="sm" onClick={() => toggleExpand(domain)}>
                                            {expandedDomain === domain ? (
                                                <FontAwesomeIcon icon={faChevronUp} />
                                            ) : (
                                                <FontAwesomeIcon icon={faChevronDown} />
                                            )}
                                        </Button>
                                    </td>
                                </tr>
                                <tr>
                                    <td colSpan="3">
                                        <Collapse isOpen={expandedDomain === domain}>
                                            <Card>
                                                <CardBody>
                                                    <Form>
                                                        <FormGroup>
                                                            <Label for="subdomainName">Subdomain Name</Label>
                                                            <Input
                                                                type="text"
                                                                id="subdomainName"
                                                                placeholder="Enter subdomain name"
                                                                value={subdomainName}
                                                                style={domainNameStyles.getInputStyles()}
                                                                onFocus={domainNameStyles.handleFocus}
                                                                onBlur={domainNameStyles.handleBlur}
                                                                onChange={(e) => setSubdomainName(e.target.value)}
                                                            />
                                                        </FormGroup>
                                                        <FormGroup>
                                                            <Label for="agentSelect">Attach Agent</Label>
                                                            <Select
                                                                name="agent"
                                                                options={agents.map((agent) => ({ value: agent.name, label: agent.name }))}
                                                                className="basic-select"
                                                                classNamePrefix="select"
                                                                styles={customStyles}
                                                                onChange={(selectedOption) => setSelectedAgent(selectedOption)}
                                                                value={selectedAgent}
                                                            />
                                                        </FormGroup>
                                                        <Row className="justify-content-end mt-3">
                                                        <Col sm="auto">
                                                        <Button
                                                            color="info" size="sm"
                                                            onClick={() => handleCreateSubdomain(domain.name)}
                                                        >
                                                            Submit
                                                        </Button>
                                                        </Col>
                                                        </Row>
                                                    </Form>
                                                    <Table className="mt-3">
                                                        <thead>
                                                            <tr>
                                                                <th>Type</th>
                                                                <th>Name</th>
                                                                <th>Data</th>
                                                                <th>Action</th>
                                                            </tr>
                                                        </thead>
                                                        <tbody>
                                                            {(domain.records || []).map((record) => (
                                                                <tr key={`${domain.name}-${record.name}`}>
                                                                    <td>{record.type}</td>
                                                                    <td>{record.name}</td>
                                                                    <td>{record.data}</td>
                                                                    <td>
                                                                        {record.type === 'A' || record.type === 'CNAME' ? (
                                                                            <FontAwesomeIcon
                                                                                icon={faTrash}
                                                                                style={{ cursor: 'pointer', color: 'lightgrey' }}
                                                                                onMouseOver={(e) => (e.target.style.color = 'red')}
                                                                                onMouseOut={(e) => (e.target.style.color = 'lightgrey')}
                                                                                onClick={() => handleDeleteSubdomain(domain.name, record.name)}
                                                                            />
                                                                        ) : null}
                                                                    </td>
                                                                </tr>
                                                            ))}
                                                        </tbody>
                                                    </Table>
                                                </CardBody>
                                            </Card>
                                        </Collapse>
                                    </td>
                                </tr>
                            </React.Fragment>
                        ))}
                    </tbody>
                </Table>
            </CardBody>
        </Card>
    );
};

export default Domains;