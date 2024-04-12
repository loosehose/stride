import React, { useMemo, useCallback } from 'react';
import Select from 'react-select';
import { Card, CardHeader, CardBody, CardTitle, FormGroup, Label, Button, Row, Col } from 'reactstrap';
import { customStyles } from './selectedStyles';
import '../assets/css/info.css';

const InfoIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" className="bi bi-info-circle" viewBox="0 0 16 16">
      <path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14zm0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16z"></path>
      <path d="m8.93 6.588-2.29.287-.082.38.45.083c.294.07.352.176.288.469l-.738 3.468c-.194.897.105 1.319.808 1.319.545 0 1.178-.252 1.465-.598l.088-.416c-.2.176-.492.246-.686.246-.275 0-.375-.193-.304-.533L8.93 6.588zM9 4.5a1 1 0 1 1-2 0 1 1 0 0 1 2 0z"></path>
  </svg>
);

const TeamserverSetup = ({ agents, selectedAgent, setSelectedAgent, selectedSoftware, setSelectedSoftware, handleSubmit }) => {
  const softwareOptions = useMemo(() => [
    { value: 'Sliver', label: 'Sliver' },
    { value: 'Mythic', label: 'Mythic | Merlin' },
    { value: 'HavocC2', label: 'Havoc C2' }
  ], []);

  const handleAgentChange = useCallback((selectedOption) => {
    setSelectedAgent(selectedOption.value);
  }, [setSelectedAgent]);

  const handleSoftwareChange = useCallback((selectedOptions) => {
    setSelectedSoftware(selectedOptions.map(option => option.value));
  }, [setSelectedSoftware]);

  return (
    <Card>
      <CardHeader><CardTitle tag="h4">Teamserver Setup</CardTitle></CardHeader>
      <CardBody>
                <div className="info-box mb-4">
                    <div className="info-icon">
                        <InfoIcon />
                    </div>
                    <div className="info-content">
                        <p><strong>Mythic Agent Requirements:</strong> 2 vCPUs, 2GB RAM</p>
                    </div>
                </div>
            </CardBody>
      <CardBody>
        <FormGroup>
          <Label>Teamserver Agent</Label>
          <Select
            options={agents}
            value={agents.find(option => option.value === selectedAgent)}
            onChange={handleAgentChange}
            styles={customStyles}
            classNamePrefix="select"
          />
        </FormGroup>
        <FormGroup>
          <Label>Software to Install</Label>
          <Select
            isMulti
            name="software"
            options={softwareOptions}
            className="basic-multi-select"
            classNamePrefix="select"
            styles={customStyles}
            onChange={handleSoftwareChange}
            value={softwareOptions.filter(option => selectedSoftware.includes(option.value))}
          />
        </FormGroup>
        <Row className="justify-content-end mt-3">
          <Col sm="auto">
            <Button color="info" size="sm" onClick={() => handleSubmit('Teamserver')}>
              Submit
            </Button>
          </Col>
        </Row>
      </CardBody>
    </Card>
  );
};

export default TeamserverSetup;