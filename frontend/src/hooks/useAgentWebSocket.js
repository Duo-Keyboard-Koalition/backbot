import { useState, useRef, useCallback } from 'react';

const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:8000/ws/run';

export function useAgentWebSocket() {
  const [status, setStatus] = useState('idle'); // idle | connecting | running | complete | timeout | error
  const [steps, setSteps] = useState([]);
  const [interventions, setInterventions] = useState([]);
  const [healthScores, setHealthScores] = useState([]);
  const [finalOutput, setFinalOutput] = useState(null);
  const [error, setError] = useState(null);
  const wsRef = useRef(null);
  const statusRef = useRef('idle');

  const setStatusBoth = (s) => {
    statusRef.current = s;
    setStatus(s);
  };

  const reset = useCallback(() => {
    setStatusBoth('idle');
    setSteps([]);
    setInterventions([]);
    setHealthScores([]);
    setFinalOutput(null);
    setError(null);
  }, []);

  const run = useCallback(
    (goal, apiKey, maxSteps = 10) => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      reset();
      setStatusBoth('connecting');

      const ws = new WebSocket(WS_URL);
      wsRef.current = ws;

      ws.onopen = () => {
        setStatusBoth('running');
        ws.send(JSON.stringify({ goal, api_key: apiKey, max_steps: maxSteps }));
      };

      ws.onmessage = (event) => {
        let msg;
        try {
          msg = JSON.parse(event.data);
        } catch {
          return;
        }

        switch (msg.type) {
          case 'start':
            setStatusBoth('running');
            break;

          case 'step_start':
            break;

          case 'step':
            setSteps((prev) => {
              const idx = prev.findIndex((s) => s.step_number === msg.step.step_number);
              if (idx >= 0) {
                const updated = [...prev];
                updated[idx] = msg.step;
                return updated;
              }
              return [...prev, msg.step];
            });
            setHealthScores((prev) => [
              ...prev,
              { step: msg.step.step_number, score: +(msg.risk_score * 10).toFixed(2) },
            ]);
            break;

          case 'intervention':
            setInterventions((prev) => [...prev, msg.intervention]);
            break;

          case 'complete':
            setSteps((prev) => {
              const idx = prev.findIndex((s) => s.step_number === msg.step.step_number);
              if (idx >= 0) {
                const updated = [...prev];
                updated[idx] = msg.step;
                return updated;
              }
              return [...prev, msg.step];
            });
            setFinalOutput(msg.step.observation);
            setStatusBoth('complete');
            break;

          case 'timeout':
            setStatusBoth('timeout');
            break;

          case 'error':
            setError(msg.message);
            setStatusBoth('error');
            break;

          case 'warning':
            console.warn('[SentinelAI]', msg.message);
            break;

          default:
            break;
        }
      };

      ws.onerror = () => {
        setError('WebSocket connection failed. Is the backend running on port 8000?');
        setStatusBoth('error');
      };

      ws.onclose = () => {
        if (statusRef.current === 'running' || statusRef.current === 'connecting') {
          setStatusBoth('idle');
        }
      };
    },
    [reset]
  );

  const stop = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
    }
    setStatusBoth('idle');
  }, []);

  return { status, steps, interventions, healthScores, finalOutput, error, run, stop, reset };
}
