import React from "react";
import { Row, Col } from "reactstrap";
import Agents from "../components/Agents";
import Domains from "../components/Domains";
import PhishingInfrastructure from "../components/PhishingInfrastructure";
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

function PhishingStudio() {

  return (
    <>
    <ToastContainer
        position="top-right"
        autoClose={5000}
        hideProgressBar={false}
        newestOnTop={false}
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        draggable
        theme="dark"
      />
      <div className="content">
        <Row>
          <Col lg="6" md="12">
            <PhishingInfrastructure />
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

export default PhishingStudio;
