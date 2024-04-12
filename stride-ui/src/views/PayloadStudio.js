import React from "react";
import { Row, Col } from "reactstrap";
import Agents from "../components/Agents";
import Domains from "../components/Domains";
import PayloadInfrastructure from "../components/PayloadInfrastructure";
import 'react-toastify/dist/ReactToastify.css';

function PayloadStudio() {

  return (
    <>
      <div className="content">
        <Row>
          <Col lg="6" md="12">
            <PayloadInfrastructure />
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

export default PayloadStudio;
