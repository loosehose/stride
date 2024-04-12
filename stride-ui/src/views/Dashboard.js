import React from "react";
import { Row, Col } from "reactstrap";
import Agents from "../components/Agents";
import Domains from "../components/Domains";
import InfrastructureDesign from "../components/Infrastructure";
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

function Dashboard() {

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
        limit={1}
      />
      <div className="content">
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

export default Dashboard;
