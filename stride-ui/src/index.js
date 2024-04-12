import React from "react";
import ReactDOM from "react-dom";
// Import HashRouter as Router
import { HashRouter as Router, Routes, Route, Navigate } from "react-router-dom";

import AdminLayout from "./layouts/Admin/Admin.js";
import Login from './views/Login/Login.js';
import { AuthProvider } from './contexts/AuthContext.js';
import { ProtectedRoute } from "./components/ProtectedRoute.js"; // Ensure this path is correct
import { WebSocketProvider } from './contexts/WebSocketContext.js';
import "./assets/scss/black-dashboard-react.scss";
import "./assets/css/nucleo-icons.css";
import "@fortawesome/fontawesome-free/css/all.min.css";

const App = () => {
  return (
    <WebSocketProvider>
      <AuthProvider>
        <Router>
          <Routes>
            <Route path="/" element={<Navigate to="/login" replace />} />
            <Route path="/login" element={<Login />} />
            <Route path="/admin/*" element={
              <ProtectedRoute>
                <AdminLayout />
              </ProtectedRoute>
            } />
          </Routes>
        </Router>
      </AuthProvider>
    </WebSocketProvider >
  );
};

const root = ReactDOM.createRoot(document.getElementById("root"));
root.render(<App />);
