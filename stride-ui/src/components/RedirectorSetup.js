import React, { useMemo, useCallback } from 'react';
import Select from 'react-select';
import { Card, CardHeader, CardBody, CardTitle, FormGroup, Label, Button, Row, Col } from 'reactstrap';
import { customStyles } from './selectedStyles';

const RedirectorSetup = ({
  agents,
  domains,
  selectedAgent,
  setSelectedAgent,
  selectedSoftware = [],
  setSelectedSoftware,
  selectedDomain,
  setSelectedDomain,
  selectedTeamserverAgent,
  setSelectedTeamserverAgent,
  handleSubmit
}) => {
  const domainOptions = useMemo(() => 
    domains.flatMap(domain => 
      domain.records.map(record => ({
        value: record.name,
        label: record.name,
      }))
    ),
    [domains]
  );

  console.log(agents)


  const softwareOptions = useMemo(() => [{ value: 'Apache', label: 'Apache' }], []);

  const handleAgentChange = useCallback((selectedOption) => {
    setSelectedAgent(selectedOption ? selectedOption.value : '');
  }, [setSelectedAgent]);

  const handleTeamserverAgentChange = useCallback((selectedOption) => {
    setSelectedTeamserverAgent(selectedOption ? selectedOption.value : '');
  }, [setSelectedTeamserverAgent]);

  const handleDomainChange = useCallback((selectedOption) => {
    setSelectedDomain(selectedOption ? selectedOption.value : '');
  }, [setSelectedDomain]);

  const handleSoftwareChange = useCallback((selectedOptions) => {
    setSelectedSoftware(selectedOptions ? selectedOptions.map(option => option.value) : []);
  }, [setSelectedSoftware]);

  return (
    <Card>
      <CardHeader><CardTitle tag="h4">HTTPS Redirector Setup</CardTitle></CardHeader>
      <CardBody>
        <FormGroup>
          <Label>Redirector Agent</Label>
          <Select
            options={agents}
            value={agents.find(option => option.value === selectedAgent)}
            onChange={handleAgentChange}
            styles={customStyles}
            classNamePrefix="select"
          />
        </FormGroup>
        <FormGroup>
          <Label>Teamserver Agent</Label>
          <Select
            options={agents}
            value={agents.find(option => option.value === selectedTeamserverAgent)}
            onChange={handleTeamserverAgentChange}
            styles={customStyles}
            classNamePrefix="select"
          />
        </FormGroup>
        <FormGroup>
          <Label>Domain</Label>
          <Select
            value={selectedDomain ? domainOptions.find(option => option.value === selectedDomain) : null}
            onChange={handleDomainChange}
            options={domainOptions}
            isClearable
            styles={customStyles}
          />
        </FormGroup>
        <FormGroup>
          <Label>Software to Install</Label>
          <Select
            isMulti
            value={softwareOptions.filter(option => selectedSoftware.includes(option.value))}
            onChange={handleSoftwareChange}
            options={softwareOptions}
            isClearable
            styles={customStyles}
          />
        </FormGroup>
        <Row className="justify-content-end mt-3">
          <Col sm="auto">
            <Button color="info" size="sm" onClick={() => handleSubmit('Redirector')}>
              Submit
            </Button>
          </Col>
        </Row>
      </CardBody>
    </Card>
  );
};

export default RedirectorSetup;