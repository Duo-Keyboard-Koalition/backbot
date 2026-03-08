import React, { useState, useEffect, useRef, useCallback } from 'react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts';
import { Activity, AlertTriangle, CheckCircle, XCircle, Shield, Zap } from 'lucide-react';

// ── design tokens ─────────────────────────────────────────────────────────────

const THEME = {
  bg: '#05050f',
  surface: 'rgba(13,13,26,0.75)',
  border: '#1e1e3a',
  text: {
    primary: '#e2e8f0',
    secondary: '#64748b',
    muted: '#334155',
  },
  risk: {
    low: '#00ff87',
    mid: '#ffb347',
    high: '#ff3860',
  },
  accent: '#3b82f6',
  font: {
    mono: "'JetBrains Mono', monospace",
    display: "'Syne', sans-serif",
  },
  radius: { sm: 6, md: 10, lg: 16, pill: 999 },
};

// ── websocket hook ─────────────────────────────────────────────────────────────

const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:8000/ws/run';

function useAgentWebSocket() {
  const [status, setStatus] = useState('idle');
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
      if (wsRef.current) wsRef.current.close();
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
        try { msg = JSON.parse(event.data); } catch { return; }

        switch (msg.type) {
          case 'start': setStatusBoth('running'); break;
          case 'step_start': break;
          case 'step':
            setSteps((prev) => {
              const idx = prev.findIndex((s) => s.step_number === msg.step.step_number);
              if (idx >= 0) { const u = [...prev]; u[idx] = msg.step; return u; }
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
              if (idx >= 0) { const u = [...prev]; u[idx] = msg.step; return u; }
              return [...prev, msg.step];
            });
            setFinalOutput(msg.step.observation);
            setStatusBoth('complete');
            break;
          case 'timeout': setStatusBoth('timeout'); break;
          case 'error': setError(msg.message); setStatusBoth('error'); break;
          case 'warning': console.warn('[SentinelAI]', msg.message); break;
          default: break;
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
    if (wsRef.current) wsRef.current.close();
    setStatusBoth('idle');
  }, []);

  return { status, steps, interventions, healthScores, finalOutput, error, run, stop, reset };
}

// ── helpers ────────────────────────────────────────────────────────────────────

function riskColor(score10) {
  if (score10 > 6) return THEME.risk.high;
  if (score10 >= 3) return THEME.risk.mid;
  return THEME.risk.low;
}

const interventionColors = {
  reprompt: THEME.accent,
  rollback: THEME.risk.mid,
  decompose: '#8b5cf6',
  halt: THEME.risk.high,
};

// ── sub-components ─────────────────────────────────────────────────────────────

function StatusBadge({ status }) {
  const cfg = {
    idle:       { icon: <Activity size={12} />,      label: 'Idle',         color: THEME.text.secondary },
    connecting: { icon: <Activity size={12} className="spin" />, label: 'Connecting…', color: THEME.accent },
    running:    { icon: <Activity size={12} className="spin" />, label: 'Running',     color: THEME.accent },
    complete:   { icon: <CheckCircle size={12} />,   label: 'Complete',     color: THEME.risk.low },
    timeout:    { icon: <AlertTriangle size={12} />, label: 'Timeout',      color: THEME.risk.mid },
    error:      { icon: <XCircle size={12} />,       label: 'Error',        color: THEME.risk.high },
  };
  const { icon, label, color } = cfg[status] ?? cfg.idle;
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 6,
      color, fontWeight: 600, fontSize: 12,
      background: color + '18',
      border: `1px solid ${color}`,
      borderRadius: THEME.radius.pill,
      padding: '4px 12px',
      fontFamily: THEME.font.mono,
    }}>
      {icon} {label}
    </span>
  );
}

function EmptyState({ icon, text }) {
  return (
    <div style={{
      border: `1px dashed ${THEME.border}`,
      borderRadius: THEME.radius.md,
      padding: 48,
      display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 16,
      color: THEME.text.muted,
      fontFamily: THEME.font.mono,
      fontSize: 12,
    }}>
      <span style={{ color: THEME.text.muted, opacity: 0.6 }}>{icon}</span>
      {text}
    </div>
  );
}

function StepCard({ step }) {
  const score10 = (step.score ?? 0) * 10;
  const isFinal = step.action === 'final_answer';
  const color = isFinal ? THEME.risk.low : riskColor(score10);
  const hasFail = step.failure_types && step.failure_types.length > 0;

  return (
    <div
      className={`slide-in${hasFail && !isFinal ? ' danger-pulse' : ''}`}
      style={{
        background: THEME.surface,
        backdropFilter: 'blur(8px)',
        borderRadius: THEME.radius.md,
        borderLeft: `3px solid ${color}`,
        border: `1px solid ${THEME.border}`,
        borderLeftWidth: 3,
        borderLeftColor: color,
        padding: 16,
        marginBottom: 8,
        fontFamily: THEME.font.mono,
      }}
    >
      {/* header row */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
        <span style={{ fontSize: 12, color: THEME.text.secondary }}>Step {step.step_number}</span>
        {isFinal ? (
          <span style={{ fontSize: 12, fontWeight: 700, color: THEME.risk.low }}>✓ Final Answer</span>
        ) : (
          <span style={{
            fontSize: 12, fontWeight: 700,
            color, fontVariantNumeric: 'tabular-nums',
            background: color + '18',
            border: `1px solid ${color}44`,
            borderRadius: THEME.radius.pill,
            padding: '2px 10px',
          }}>
            Risk {score10.toFixed(1)}
          </span>
        )}
      </div>

      {/* action — medium weight */}
      <div style={{ fontSize: 14, fontWeight: 600, color: THEME.text.primary, marginBottom: 8 }}>
        <code style={{
          background: THEME.border + '66',
          borderRadius: THEME.radius.sm,
          padding: '2px 8px',
          fontSize: 12,
        }}>
          {step.action}
        </code>
        {step.action_input && Object.keys(step.action_input).length > 0 && (
          <span style={{ color: THEME.text.secondary, fontSize: 12, marginLeft: 8 }}>
            {JSON.stringify(step.action_input)}
          </span>
        )}
      </div>

      {/* thought — lower weight */}
      <div style={{ fontSize: 12, fontWeight: 400, color: THEME.text.secondary, marginBottom: 8 }}>
        <span style={{ color: THEME.text.muted }}>Thought: </span>{step.thought}
      </div>

      {/* observation — lowest weight */}
      <div style={{
        fontSize: 12, fontWeight: 400, color: THEME.text.muted,
        background: THEME.bg,
        borderRadius: THEME.radius.sm,
        padding: 8,
      }}>
        <span style={{ color: THEME.text.muted }}>Observation: </span>{step.observation}
      </div>

      {/* failure pills */}
      {hasFail && (
        <div style={{ marginTop: 8, display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          {step.failure_types.map((f) => (
            <span key={f} style={{
              fontSize: 12, fontWeight: 600,
              color: THEME.risk.high,
              background: THEME.risk.high + '18',
              border: `1px solid ${THEME.risk.high}44`,
              borderRadius: THEME.radius.pill,
              padding: '2px 8px',
            }}>
              {f}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

function InterventionBar({ intervention }) {
  const color = interventionColors[intervention.intervention_type] ?? THEME.text.secondary;
  const preview = intervention.reprompt ?? '';
  return (
    <div
      className="slide-in"
      style={{
        background: THEME.surface,
        backdropFilter: 'blur(8px)',
        borderRadius: THEME.radius.md,
        border: `1px solid ${THEME.border}`,
        borderLeft: `4px solid ${color}`,
        padding: 16,
        marginBottom: 8,
        fontFamily: THEME.font.mono,
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
        <span style={{
          fontSize: 12, fontWeight: 700,
          color, textTransform: 'uppercase', letterSpacing: '0.06em',
        }}>
          {intervention.intervention_type}
        </span>
        <span style={{ fontSize: 12, color: THEME.text.muted }}>Step {intervention.step_number}</span>
      </div>
      <div style={{ fontSize: 12, color: THEME.text.secondary, marginBottom: preview ? 8 : 0 }}>
        {intervention.reason}
      </div>
      {preview && (
        <div style={{ fontSize: 12, color: THEME.text.muted, fontStyle: 'italic' }}>
          "{preview.length > 160 ? preview.slice(0, 160) + '…' : preview}"
        </div>
      )}
    </div>
  );
}

const CustomTooltip = ({ active, payload, label }) => {
  if (!active || !payload || !payload.length) return null;
  return (
    <div style={{
      background: THEME.surface,
      border: `1px solid ${THEME.border}`,
      borderRadius: THEME.radius.sm,
      padding: '8px 12px',
      fontFamily: THEME.font.mono,
      fontSize: 12,
    }}>
      <div style={{ color: THEME.text.secondary, marginBottom: 4 }}>Step {label}</div>
      <div style={{ color: THEME.accent, fontWeight: 700 }}>
        Risk: {payload[0].value.toFixed(2)}
      </div>
    </div>
  );
};

// ── main app ───────────────────────────────────────────────────────────────────

export default function App() {
  const [goal, setGoal] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [maxSteps, setMaxSteps] = useState(10);
  const [goalFocused, setGoalFocused] = useState(false);
  const prevInterventionCount = useRef(0);
  const [counterAnim, setCounterAnim] = useState(false);

  const { status, steps, interventions, healthScores, finalOutput, error, run, stop } =
    useAgentWebSocket();

  const isRunning = status === 'connecting' || status === 'running';

  // inject Google Fonts
  useEffect(() => {
    const link = document.createElement('link');
    link.href =
      'https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600;700&family=Syne:wght@700;800&display=swap';
    link.rel = 'stylesheet';
    document.head.appendChild(link);
  }, []);

  // inject global styles
  useEffect(() => {
    const style = document.createElement('style');
    style.textContent = `
      * { box-sizing: border-box; margin: 0; padding: 0; }
      body { background: #05050f; color: #e2e8f0; font-family: 'JetBrains Mono', monospace; }
      ::-webkit-scrollbar { width: 4px; }
      ::-webkit-scrollbar-thumb { background: #1e1e3a; border-radius: 4px; }
      @keyframes spin { to { transform: rotate(360deg); } }
      @keyframes slideIn {
        from { opacity: 0; transform: translateY(12px); }
        to   { opacity: 1; transform: translateY(0); }
      }
      @keyframes dangerPulse {
        0%, 100% { box-shadow: 0 0 0px 0px rgba(255,56,96,0); }
        50%       { box-shadow: 0 0 14px 3px rgba(255,56,96,0.35); }
      }
      @keyframes interventionPop {
        0%   { transform: scale(1); }
        50%  { transform: scale(1.3); }
        100% { transform: scale(1); }
      }
      .spin { animation: spin 1s linear infinite; }
      .slide-in { animation: slideIn 0.2s ease forwards; }
      .danger-pulse { animation: dangerPulse 2s ease-in-out infinite; }
      .intervention-pop { animation: interventionPop 0.3s ease; }
    `;
    document.head.appendChild(style);
  }, []);

  // animate intervention counter on increment
  useEffect(() => {
    if (interventions.length > prevInterventionCount.current) {
      setCounterAnim(false);
      requestAnimationFrame(() => {
        requestAnimationFrame(() => setCounterAnim(true));
      });
      setTimeout(() => setCounterAnim(false), 350);
    }
    prevInterventionCount.current = interventions.length;
  }, [interventions.length]);

  const handleRun = () => {
    if (!goal.trim()) return;
    run(goal.trim(), apiKey.trim(), maxSteps);
  };

  // shared input style
  const inputBase = {
    width: '100%',
    background: THEME.bg,
    color: THEME.text.primary,
    border: `1.5px solid ${THEME.border}`,
    borderRadius: THEME.radius.md,
    padding: '8px 16px',
    fontSize: 14,
    fontFamily: THEME.font.mono,
    outline: 'none',
    transition: 'all 0.15s ease',
  };

  return (
    <div style={{ minHeight: '100vh', background: THEME.bg }}>

      {/* ── header ── */}
      <header style={{
        background: THEME.bg,
        borderBottom: `1px solid ${THEME.border}`,
        backgroundImage: 'linear-gradient(180deg, #0d0d1a 0%, transparent 100%)',
        height: 56,
        padding: '0 32px',
        display: 'flex', alignItems: 'center', gap: 16,
        position: 'sticky', top: 0, zIndex: 100,
      }}>
        <Shield size={20} color={THEME.accent} />
        <span style={{
          fontFamily: THEME.font.display,
          fontSize: 18,
          fontWeight: 800,
          background: 'linear-gradient(90deg, #00ff87, #3b82f6)',
          WebkitBackgroundClip: 'text',
          WebkitTextFillColor: 'transparent',
        }}>
          SentinelAI
        </span>
        <span style={{ marginLeft: 'auto' }}>
          <StatusBadge status={status} />
        </span>
      </header>

      <div style={{ maxWidth: 1280, margin: '0 auto', padding: '32px 24px' }}>

        {/* ── control panel ── */}
        <div style={{
          background: THEME.surface,
          backdropFilter: 'blur(12px)',
          border: `1px solid ${THEME.border}`,
          borderRadius: THEME.radius.lg,
          padding: 24,
          marginBottom: 24,
          display: 'flex', gap: 16, flexWrap: 'wrap', alignItems: 'flex-end',
        }}>
          {/* goal */}
          <div style={{ flex: '2 1 320px' }}>
            <label style={{ fontSize: 12, color: THEME.text.secondary, display: 'block', marginBottom: 8, fontFamily: THEME.font.mono }}>
              Goal
            </label>
            <input
              value={goal}
              onChange={(e) => setGoal(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && !isRunning && handleRun()}
              onFocus={() => setGoalFocused(true)}
              onBlur={() => setGoalFocused(false)}
              placeholder="What should the agent accomplish?"
              disabled={isRunning}
              style={{
                ...inputBase,
                borderRadius: THEME.radius.pill,
                borderColor: goalFocused ? THEME.accent : THEME.border,
                boxShadow: goalFocused ? `0 0 12px 2px ${THEME.accent}44` : 'none',
              }}
            />
          </div>

          {/* api key */}
          <div style={{ flex: '1 1 200px' }}>
            <label style={{ fontSize: 12, color: THEME.text.secondary, display: 'block', marginBottom: 8, fontFamily: THEME.font.mono }}>
              Gemini API Key
            </label>
            <input
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              type="password"
              placeholder="AIza…"
              disabled={isRunning}
              style={inputBase}
            />
          </div>

          {/* max steps */}
          <div style={{ flex: '0 0 96px' }}>
            <label style={{ fontSize: 12, color: THEME.text.secondary, display: 'block', marginBottom: 8, fontFamily: THEME.font.mono }}>
              Max Steps
            </label>
            <input
              value={maxSteps}
              onChange={(e) => setMaxSteps(Number(e.target.value))}
              type="number"
              min={1}
              max={20}
              disabled={isRunning}
              style={inputBase}
            />
          </div>

          {/* buttons */}
          <div style={{ display: 'flex', gap: 8 }}>
            <button
              onClick={handleRun}
              disabled={isRunning || !goal.trim()}
              style={{
                background: isRunning || !goal.trim() ? THEME.text.muted : THEME.accent,
                color: THEME.text.primary,
                border: 'none',
                borderRadius: THEME.radius.pill,
                padding: '8px 24px',
                fontWeight: 700, fontSize: 14,
                fontFamily: THEME.font.mono,
                cursor: isRunning || !goal.trim() ? 'not-allowed' : 'pointer',
                transition: 'all 0.15s ease',
                filter: 'brightness(1)',
              }}
              onMouseEnter={(e) => { if (!isRunning && goal.trim()) e.currentTarget.style.filter = 'brightness(1.15)'; }}
              onMouseLeave={(e) => { e.currentTarget.style.filter = 'brightness(1)'; }}
            >
              Run
            </button>
            {isRunning && (
              <button
                onClick={stop}
                style={{
                  background: THEME.risk.high,
                  color: THEME.text.primary,
                  border: 'none',
                  borderRadius: THEME.radius.pill,
                  padding: '8px 24px',
                  fontWeight: 700, fontSize: 14,
                  fontFamily: THEME.font.mono,
                  cursor: 'pointer',
                  transition: 'all 0.15s ease',
                  filter: 'brightness(1)',
                }}
                onMouseEnter={(e) => { e.currentTarget.style.filter = 'brightness(1.15)'; }}
                onMouseLeave={(e) => { e.currentTarget.style.filter = 'brightness(1)'; }}
              >
                Stop
              </button>
            )}
          </div>
        </div>

        {/* ── error banner ── */}
        {error && (
          <div style={{
            background: THEME.risk.high + '18',
            border: `1px solid ${THEME.risk.high}44`,
            color: THEME.risk.high,
            borderRadius: THEME.radius.md,
            padding: '12px 16px',
            marginBottom: 16,
            fontSize: 14,
            fontFamily: THEME.font.mono,
          }}>
            <strong>Error:</strong> {error}
          </div>
        )}

        {/* ── final answer banner ── */}
        {finalOutput && (
          <div style={{
            background: THEME.risk.low + '0f',
            border: `1px solid ${THEME.risk.low}44`,
            borderRadius: THEME.radius.md,
            padding: '16px 24px',
            marginBottom: 24,
            fontFamily: THEME.font.mono,
          }}>
            <div style={{ fontWeight: 700, color: THEME.risk.low, fontSize: 12, marginBottom: 8 }}>
              FINAL ANSWER
            </div>
            <div style={{ fontSize: 14, color: THEME.text.primary, whiteSpace: 'pre-wrap' }}>
              {finalOutput}
            </div>
          </div>
        )}

        {/* ── two-column layout ── */}
        <div style={{ display: 'flex', gap: 24, alignItems: 'flex-start' }}>

          {/* left column */}
          <div style={{ flex: '1 1 0', minWidth: 0 }}>

            {/* risk score chart */}
            <div style={{
              background: THEME.surface,
              backdropFilter: 'blur(8px)',
              border: `1px solid ${THEME.border}`,
              borderRadius: THEME.radius.lg,
              padding: 24,
              marginBottom: 24,
            }}>
              <div style={{
                fontWeight: 700, fontSize: 14,
                color: THEME.text.primary,
                fontFamily: THEME.font.mono,
                marginBottom: 16,
              }}>
                Agent Risk Score (0 – 10)
              </div>

              {healthScores.length === 0 ? (
                <EmptyState icon={<Activity size={32} />} text="No data — run the agent to see live risk scores." />
              ) : (
                <ResponsiveContainer width="100%" height={216}>
                  <AreaChart data={healthScores} margin={{ top: 8, right: 16, left: -16, bottom: 8 }}>
                    <defs>
                      <linearGradient id="riskGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="0%" stopColor={THEME.accent} stopOpacity={0.6} />
                        <stop offset="100%" stopColor={THEME.accent} stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke={THEME.border} />
                    <XAxis
                      dataKey="step"
                      tick={{ fill: THEME.text.secondary, fontSize: 11, fontFamily: THEME.font.mono }}
                      axisLine={{ stroke: THEME.border }}
                      tickLine={false}
                    />
                    <YAxis
                      domain={[0, 10]}
                      tick={{ fill: THEME.text.secondary, fontSize: 11, fontFamily: THEME.font.mono }}
                      axisLine={{ stroke: THEME.border }}
                      tickLine={false}
                    />
                    <Tooltip content={<CustomTooltip />} />
                    <ReferenceLine
                      y={5}
                      stroke={THEME.risk.high}
                      strokeDasharray="5 3"
                      label={{
                        value: 'Threshold',
                        fill: THEME.risk.high,
                        fontSize: 11,
                        fontFamily: THEME.font.mono,
                        position: 'insideTopRight',
                      }}
                    />
                    <Area
                      type="monotone"
                      dataKey="score"
                      stroke={THEME.accent}
                      strokeWidth={2}
                      fill="url(#riskGradient)"
                      dot={{ r: 4, fill: THEME.accent, strokeWidth: 0 }}
                      activeDot={{ r: 6, fill: THEME.accent }}
                      isAnimationActive={false}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              )}
            </div>

            {/* step cards */}
            <div>
              <div style={{
                fontWeight: 700, fontSize: 14,
                color: THEME.text.primary,
                fontFamily: THEME.font.mono,
                marginBottom: 16,
              }}>
                Steps{' '}
                {steps.length > 0 && (
                  <span style={{ color: THEME.text.muted, fontWeight: 400 }}>({steps.length})</span>
                )}
              </div>
              {steps.length === 0 ? (
                <EmptyState icon={<Zap size={32} />} text="Agent steps will appear here as they complete." />
              ) : (
                steps.map((step) => <StepCard key={step.step_number} step={step} />)
              )}
            </div>
          </div>

          {/* right sidebar — interventions */}
          <div style={{ width: 320, flexShrink: 0 }}>
            <div style={{
              background: THEME.surface,
              backdropFilter: 'blur(8px)',
              border: `1px solid ${THEME.border}`,
              borderRadius: THEME.radius.lg,
              padding: 24,
              position: 'sticky',
              top: 72,
            }}>
              <div style={{
                fontWeight: 700, fontSize: 14,
                color: THEME.text.primary,
                fontFamily: THEME.font.mono,
                marginBottom: 16,
                display: 'flex', alignItems: 'center', gap: 8,
              }}>
                <AlertTriangle size={14} color={THEME.risk.mid} />
                Interventions
                {interventions.length > 0 && (
                  <span
                    className={counterAnim ? 'intervention-pop' : ''}
                    style={{
                      marginLeft: 'auto',
                      background: THEME.risk.mid + '22',
                      color: THEME.risk.mid,
                      border: `1px solid ${THEME.risk.mid}44`,
                      borderRadius: THEME.radius.pill,
                      padding: '2px 10px',
                      fontSize: 12,
                      fontWeight: 700,
                    }}
                  >
                    {interventions.length}
                  </span>
                )}
              </div>

              {interventions.length === 0 ? (
                <EmptyState icon={<Shield size={28} />} text="No interventions yet." />
              ) : (
                interventions.map((iv, i) => <InterventionBar key={i} intervention={iv} />)
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
