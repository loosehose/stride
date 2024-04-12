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
} from 'reactstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faCircleInfo } from '@fortawesome/free-solid-svg-icons';
import '../assets/css/info.css';
import { customStyles } from './selectedStyles';
import { useInputStyles } from './inputStyles';

const InfoIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" className="bi bi-info-circle" viewBox="0 0 16 16">
        <path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14zm0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16z"></path>
        <path d="m8.93 6.588-2.29.287-.082.38.45.083c.294.07.352.176.288.469l-.738 3.468c-.194.897.105 1.319.808 1.319.545 0 1.178-.252 1.465-.598l.088-.416c-.2.176-.492.246-.686.246-.275 0-.375-.193-.304-.533L8.93 6.588zM9 4.5a1 1 0 1 1-2 0 1 1 0 0 1 2 0z"></path>
    </svg>
  );

// Static options moved outside the component
const processOptions = ['explorer.exe', 'notepad.exe'].map(p => ({ value: p, label: p }));
const methodOptions = [
    'PoolPartyModuleStomping', 'PoolParty', 'ThreadlessInject', 'ModuleStomping', 'QueueUserAPC',
    'ProcessHollow', 'EnumDisplayMonitors', 'RemoteThreadContext', 'RemoteThreadSuspended', 'CurrentThread'
].map(m => ({ value: m, label: m }));
const syscallOptions = ['SysWhispers2', 'SysWhispers3', 'GetSyscallStub', 'None'].map(s => ({ value: s, label: s }));
const booleanOptions = [
    { value: true, label: 'Yes' },
    { value: false, label: 'No' }
];

const PayloadSetup = ({ handleSubmit, agents }) => {

    const [selectedMethod, setSelectedMethod] = useState('');
    const [selectedSyscall, setSelectedSyscall] = useState('');

    const [sandboxArg, setSandboxArg] = useState('');

    const [outfile, setOutfile] = useState('');
    const [ppidPriv, setPpidPriv] = useState(false);
    const [targetDll, setTargetDll] = useState('');
    const [exportFunction, setExportFunction] = useState('');

    const [formData, setFormData] = useState({
        selectedAgent: '',
        shellcodePath: '',
        selectedProcess: '',
        selectedMethod: '',
        selectedSyscall: '', 
        unhookNtdll: true,
        noRandomize: false,
        noSandbox: false,
        dll: false,
        dllProxy: '',
        outfile: '',
        createProcess: false,
        targetDll: '',
        exportFunction: '',
    });

    const shellcodePathInputStyles = useInputStyles();
    const sandboxArgInputStyles = useInputStyles();
    const outfileInputStyles = useInputStyles();

    const agentOptions = agents.map(agent => ({ value: agent.ip, label: agent.name }));

    // Handler to update consolidated state
    const handleInputChange = (field, value) => {
        setFormData(prevState => ({
            ...prevState,
            [field]: value
        }));
    };

    const handleSubmitInternal = () => {
        const payload = {
            agentIP: formData.selectedAgent?.value ?? '',
            shellcodePath: formData.shellcodePath,
            process: formData.selectedProcess?.value ?? '',
            method: formData.selectedMethod?.value ?? '',
            unhookNtdll: formData.unhookNtdll?.value ?? false,
            syscall: formData.selectedSyscall?.value ?? '',
            dll: formData.dll?.value ?? false,
            outfile: outfile,
            sandboxArg: sandboxArg,
        };
    
        handleSubmit(payload);
    };
    return (
        <Card>
            <CardHeader>
                <CardTitle tag="h4">Payload Setup</CardTitle>
            </CardHeader>
            <CardBody>
                <div className="info-box mb-4">
                    <div className="info-icon">
                        <InfoIcon />
                    </div>
                    <div className="info-content">
                    <p><strong>Shellcode (.bin) file:</strong> Must be stored on Agent</p>
                    <p><strong>Agent Requirements:</strong> 2 vCPUs, 4GB RAM</p>
                    </div>
                </div>
            </CardBody>
            <CardBody>
            <FormGroup>
                    <Label>Shellcode Path</Label>
                    <Input
                        type="text"
                        value={formData.shellcodePath}
                        onChange={(e) => handleInputChange('shellcodePath', e.target.value)}
                        placeholder='/path/to/shellcode.bin'
                        style={shellcodePathInputStyles.getInputStyles()}
                        onFocus={shellcodePathInputStyles.handleFocus}
                        onBlur={shellcodePathInputStyles.handleBlur}
                    />
                </FormGroup>

                <FormGroup>
                    <Label>Process to Inject Into</Label>
                    <Select
                        options={processOptions}
                        value={formData.selectedProcess}
                        onChange={(value) => handleInputChange('selectedProcess', value)}
                        classNamePrefix="select"
                        styles={customStyles}
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Method for Shellcode Execution</Label>
                    <Select
                        options={methodOptions}
                        value={formData.selectedMethod}
                        onChange={(value) => handleInputChange('selectedMethod', value)}
                        classNamePrefix="select"
                        styles={customStyles}
                    />
                </FormGroup>
                {selectedMethod && selectedMethod.value === 'ThreadlessInject' && (
                    <>
                        <FormGroup>
                            <Label>Create Process</Label>
                            <Select
                                options={booleanOptions}
                                value={ppidPriv}
                                className="basic-single"
                                classNamePrefix="select"
                                styles={customStyles}
                                onChange={(value) => handleInputChange('ppidPriv', value)}
                            />
                        </FormGroup>
                        <FormGroup>
                            <Label>Target DLL</Label>
                            <Input
                                type="text"
                                value={targetDll}
                                style={customStyles}
                                placeholder='ntdll.dll'
                                onChange={(value) => setTargetDll('dll', value)}
                            />
                        </FormGroup>
                        <FormGroup>
                            <Label>Export Function</Label>
                            <Input
                                type="text"
                                value={exportFunction}
                                style={customStyles}
                                placeholder='NtClose'
                                onChange={(e) => setExportFunction(e.target.value)}
                            />
                        </FormGroup>
                    </>
                )}
                <FormGroup>
                    <Label>Unhook NTDLL</Label>
                    <Select
                        options={booleanOptions}
                        value={formData.unhookNtdll}
                        onChange={(value) => handleInputChange('unhookNtdll', value)}
                        classNamePrefix="select"
                        styles={customStyles}
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Sandbox Argument</Label>
                    <Input
                        type="text"
                        value={sandboxArg}
                        onChange={(e) => setSandboxArg(e.target.value)}
                        style={sandboxArgInputStyles.getInputStyles()}
                        onFocus={sandboxArgInputStyles.handleFocus}
                        onBlur={sandboxArgInputStyles.handleBlur}
                        placeholder='domain.local, machine.domain.local, username'
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Syscall Execution Method</Label>
                    <Select
                        options={syscallOptions}
                        value={selectedSyscall} // Make sure this is the state you're updating
                        onChange={(option) => setSelectedSyscall(option)} // Directly update the state with the selected option
                        classNamePrefix="select"
                        styles={customStyles}
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Generate DLL</Label>
                    <Select
                        options={booleanOptions}
                        value={formData.dll}
                        onChange={(value) => handleInputChange('dll', value)}
                        className="basic-single"
                        classNamePrefix="select"
                        styles={customStyles}
                    />
                </FormGroup>

                <FormGroup>
                    <Label for="outfile">Output File</Label>
                    <Input
                        type="text"
                        id="outfile"
                        value={outfile}
                        onChange={(e) => setOutfile(e.target.value)}
                        style={outfileInputStyles.getInputStyles()}
                        onFocus={outfileInputStyles.handleFocus}
                        onBlur={outfileInputStyles.handleBlur}
                    />
                </FormGroup>
                <FormGroup>
                    <Label>Select Agent</Label>
                    <Select
                        options={agentOptions}
                        value={formData.selectedAgent}
                        onChange={(value) => handleInputChange('selectedAgent', value)}
                        classNamePrefix="select"
                        styles={customStyles}
                    />
                </FormGroup>

                <Button  color="info" size="sm" onClick={handleSubmitInternal}>Submit Payload Setup</Button>
            </CardBody>
        </Card>
    );
}
export default PayloadSetup;