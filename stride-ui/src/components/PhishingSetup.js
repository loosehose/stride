import React, { useState } from 'react';
import Select from 'react-select';
import {
    Card,
    CardHeader,
    CardBody,
    CardTitle,
    FormGroup,
    Label,
    Input,
    Button,
    Row,
    Col,
} from 'reactstrap';
import '../assets/css/info.css';
import { customStyles } from './selectedStyles';
import { useInputStyles } from './inputStyles';

const InfoIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" className="bi bi-info-circle" viewBox="0 0 16 16">
        <path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14zm0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16z"></path>
        <path d="m8.93 6.588-2.29.287-.082.38.45.083c.294.07.352.176.288.469l-.738 3.468c-.194.897.105 1.319.808 1.319.545 0 1.178-.252 1.465-.598l.088-.416c-.2.176-.492.246-.686.246-.275 0-.375-.193-.304-.533L8.93 6.588zM9 4.5a1 1 0 1 1-2 0 1 1 0 0 1 2 0z"></path>
    </svg>
  );

const PhishingSetup = ({ agents, domains, handleSubmit }) => {
    const [selectedAgent, setSelectedAgent] = useState('');
    const [selectedRootDomain, setSelectedRootDomain] = useState('');
    const [subdomains, setSubdomains] = useState('');
    const [rootDomainBool, setRootDomainBool] = useState('false');
    const [redirectUrl, setRedirectUrl] = useState('');
    const [feedBool, setFeedBool] = useState('true');
    const [ridReplacement, setRidReplacement] = useState('');
    const [blacklistBool, setBlacklistBool] = useState('false');

    const agentOptions = agents.map(agent => ({ value: agent.ip, label: agent.name }));
    const domainOptions = domains.map(domain => ({ value: domain.name, label: domain.name }));

    const subdomainNameStyles = useInputStyles();
    const redirectUrlStyles = useInputStyles();
    const ridReplacementStyles = useInputStyles();

    const booleanOptions = [
        { value: 'true', label: 'True' },
        { value: 'false', label: 'False' }
    ];

    const handleAgentChange = (selectedOption) => {
        setSelectedAgent(selectedOption);
    };

    const handleRootDomainChange = (selectedOption) => {
        setSelectedRootDomain(selectedOption);
    };

    return (
        <Card>
            <CardHeader>
                <CardTitle tag="h4">Phishing Setup</CardTitle>
            </CardHeader>
            <CardBody>
                <div className="info-box mb-4">
                    <div className="info-icon">
                        <InfoIcon />
                    </div>
                    <div className="info-content">
                    <p><strong>To access GoPhish: </strong><code>ssh -L 3333:localhost:3333 root@agent-ip</code></p>
                    </div>
                </div>
            </CardBody>
            <CardBody>
                <FormGroup>
                    <Label>Select Agent</Label>
                    <Select
                        options={agentOptions}
                        value={selectedAgent}
                        onChange={handleAgentChange}
                        styles={customStyles}
                        classNamePrefix="select"
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Select Root Domain</Label>
                    <Select
                        options={domainOptions}
                        value={selectedRootDomain}
                        onChange={handleRootDomainChange}
                        styles={customStyles}
                        classNamePrefix="select"
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Subdomains (space separated)</Label>
                    <Input
                        type="text"
                        value={subdomains}
                        onChange={e => setSubdomains(e.target.value)}
                        placeholder="www account"
                        style={subdomainNameStyles.getInputStyles()}
                        onFocus={subdomainNameStyles.handleFocus}
                        onBlur={subdomainNameStyles.handleBlur}
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Proxy Root Domain to EvilGoPhish</Label>
                    <Select
                        options={booleanOptions}
                        value={booleanOptions.find(option => option.value === rootDomainBool)}
                        onChange={selectedOption => setRootDomainBool(selectedOption.value)}
                        styles={customStyles}
                        classNamePrefix="select"
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Redirect URL</Label>
                    <Input
                        type="text"
                        value={redirectUrl}
                        onChange={e => setRedirectUrl(e.target.value)}
                        placeholder='https://www.google.com/'
                        style={redirectUrlStyles.getInputStyles()}
                        onFocus={redirectUrlStyles.handleFocus}
                        onBlur={redirectUrlStyles.handleBlur}
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Use Live Feed</Label>
                    <Select
                        options={booleanOptions}
                        value={booleanOptions.find(option => option.value === feedBool)}
                        onChange={selectedOption => setFeedBool(selectedOption.value)}
                        styles={customStyles}
                        classNamePrefix="select"
                    />
                </FormGroup>
                <FormGroup>
                    <Label>RID Replacement</Label>
                    <Input
                        type="text"
                        value={ridReplacement}
                        onChange={e => setRidReplacement(e.target.value)}
                        placeholder='user_id'
                        style={ridReplacementStyles.getInputStyles()}
                        onFocus={ridReplacementStyles.handleFocus}
                        onBlur={ridReplacementStyles.handleBlur}
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Use Apache Blacklist</Label>
                    <Select
                        options={booleanOptions}
                        value={booleanOptions.find(option => option.value === blacklistBool)}
                        onChange={selectedOption => setBlacklistBool(selectedOption.value)}
                        styles={customStyles}
                        classNamePrefix="select"
                    />
                </FormGroup>
                <Row className="justify-content-end mt-3">
                    <Col sm="auto">
                        <Button color="info" size="sm" onClick={() => handleSubmit({
                            agentIP: selectedAgent.value,
                            rootDomain: selectedRootDomain.value,
                            subdomains: subdomains.split(' '),
                            rootDomainBool,
                            redirectUrl,
                            feedBool,
                            ridReplacement,
                            blacklistBool,
                        })}>Submit Phishing Setup</Button>
                    </Col>
                </Row>
            </CardBody>
        </Card>
    );
};

export default PhishingSetup;