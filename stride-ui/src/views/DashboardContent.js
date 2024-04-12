import React from "react";
import { Row, Col, Alert } from "reactstrap";
import { useNotifications } from "./NotificationContext"; // Adjust the import path as necessary
import Agents from "../components/Agents"; // Update paths as necessary
import Domains from "../components/Domains"; // Update paths as necessary
import InfrastructureDesign from "../components/Infrastructure"; // Update paths as necessary

function DashboardContent() {
    const { notifications } = useNotifications(); // This line is crucial

    return (
        <>
            <div className="content">
                {notifications.map((notification, index) => (
                    <div key={index} className={`alert alert-${notification.type}`} role="alert">
                        {notification.message}
                    </div>
                ))}
                <Row>
                    <Col lg="6" md="12">
                        <InfrastructureDesign />
                    </Col>
                    <Col lg="6" md="12">
                        <Agents />
                        <Domains />
                    </Col>
                </Row>
            </div>
        </>
    );
}

export default DashboardContent; // Make sure to export DashboardContent
