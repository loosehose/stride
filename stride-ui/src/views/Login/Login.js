import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { Card, CardHeader, CardBody, FormGroup, Label, Input, Button, Container, Row, Col } from 'reactstrap';
import '../../assets/scss/black-dashboard-react.scss'; // Ensure this is the correct path to your CSS file

const Login = () => {
    const [apiKey, setApiKey] = useState('');
    const { login, loading, error, isAuthenticated } = useAuth();
    const navigate = useNavigate();

    const handleLogin = async (e) => {
        e.preventDefault();
        await login(apiKey);
    };

    useEffect(() => {
        if (isAuthenticated) {
            navigate('/admin/dashboard');
        }
    }, [isAuthenticated, navigate]);

    return (
        <Container className="vh-100">
            <Row className="justify-content-center align-items-center h-100">
                <Col lg="4" md="6">
                    <Card className={`mt-5 ${loading || error ? 'shake' : ''}`}>
                        <CardHeader>
                            <h4 className="card-title">Login</h4>
                        </CardHeader>
                        <CardBody>
                            <form onSubmit={handleLogin}>
                                <FormGroup>
                                    <Label for="apiKey">DigitalOcean API Key</Label>
                                    <Input
                                        type="text"
                                        name="apiKey"
                                        id="apiKey"
                                        placeholder="Enter your API key"
                                        value={apiKey}
                                        onChange={(e) => setApiKey(e.target.value)}
                                        disabled={loading}
                                        required
                                    />
                                </FormGroup>
                                <div className="text-center">
                                    <Button color="info" size="sm" type="submit" disabled={loading}>
                                        {loading ? 'Loading...' : 'Login'}
                                    </Button>
                                </div>
                                {error && <div className="alert alert-danger mt-3">{error}</div>}
                            </form>
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        </Container>
    );
};

export default Login;
