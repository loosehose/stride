import { useCallback, useRef } from 'react';
import { toast } from 'react-toastify';
import { NotificationConfig } from '../components/notificationConfig';

export const useNotifications = () => {
  const toastId = useRef(null);
  const lastMessage = useRef('');
  const lastReceivedMessage = useRef('');
  const lastMessageId = useRef('');

  const handleWebSocketMessage = useCallback((event) => {
    let data;
    try {
      data = JSON.parse(event.data);
    } catch (error) {
      console.error('Received invalid WebSocket message:', event.data);
      return;
    }

    const { level, message, id } = data;

    // Check if the received message has the same ID as the last message
    if (id && id === lastMessageId.current) {
      return;
    }
    lastMessageId.current = id;

    // Check if the received message is the same as the last received message
    if (message === lastReceivedMessage.current) {
      return;
    }
    lastReceivedMessage.current = message;

    if (level === 'EXEC') {
      if (!toastId.current || !toast.isActive(toastId.current)) {
        toastId.current = toast.loading(message, { ...NotificationConfig, toastId: 'exec' });
        lastMessage.current = message;
      } else {
        if (message !== lastMessage.current) {
          toast.update(toastId.current, { render: message, ...NotificationConfig });
          lastMessage.current = message;
        }
      }
    } else if (level === 'SUCCESS' || level === 'ERROR') {
      if (toastId.current && toast.isActive(toastId.current)) {
        toast.dismiss(toastId.current);
      }
      toastId.current = null;

      if (message !== lastMessage.current) {
        const toastFunc = level === 'SUCCESS' ? toast.success : toast.error;
        toastFunc(message, { ...NotificationConfig });
        lastMessage.current = message;
      }
    } else {
      if (message !== lastMessage.current) {
        const toastFunc = level === 'INFO' ? toast.info : toast.warn;
        toastFunc(message, { ...NotificationConfig });
        lastMessage.current = message;
      }
    }
  }, []);

  return { handleWebSocketMessage };
};