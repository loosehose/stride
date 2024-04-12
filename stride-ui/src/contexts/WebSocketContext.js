import React, { createContext, useContext, useEffect, useState } from 'react';
import { useNotifications } from '../hooks/useNotifications';

const WebSocketContext = createContext();

export const useWebSocket = () => useContext(WebSocketContext);

export const WebSocketProvider = ({ children }) => {
  const [ws, setWs] = useState(null);
  const { handleWebSocketMessage } = useNotifications();

  useEffect(() => {
    const setupWebSocket = () => {
      const websocket = new WebSocket('ws://localhost:8080/ws');
      setWs(websocket);

      websocket.onopen = () => console.log('WebSocket connection established');
      websocket.onmessage = handleWebSocketMessage;
      websocket.onclose = () => {
        console.log('WebSocket connection closed');
        setTimeout(setupWebSocket, 5000);
      };
      websocket.onerror = (error) => console.error('WebSocket error:', error);
    };

    setupWebSocket();

    return () => ws?.close();
  }, [handleWebSocketMessage]);

  return (
    <WebSocketContext.Provider value={ws}>
      {children}
    </WebSocketContext.Provider>
  );
};