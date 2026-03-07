import React, { useState } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts';
import { Activity, AlertTriangle, CheckCircle, XCircle, Shield } from 'lucide-react';
import { useAgentWebSocket } from './hooks/useAgentWebSocket';
import './App.css';

// ── helpers ──────────────────────────────────────────────────────────────────

function riskColor(score10) {
  if (score10 >= 7) return '#ef4444';
  if (score10 >= 4) return '#f59e0b';
  return '#22c55e';
}

// ── sub-components ────────────────────────────────────────────────────────────

function StatusBadge({ status }) {
  const cfg = {
    idle:       { icon: <Activity size={13} />,      label: 'Idle',          color: '#6b7280' },
    connecting: { icon: <Activity size={13} className="spin" />, label: 'Connecting…', color: '#3b82f6' },
    running:    { icon: <Activity size={13} className="spin" />, label: 'Running',     color: '#3b82f6' },
    complete:   { icon: <CheckCircle size={13} />,   label: 'Complete',      color: '#22c55e' },
    timeout:    { icon: <AlertTriangle size={13} />, label: 'Timeout',       color: '#f59e0b' },
    error:      { icon: <XCircle size={13} />,       label: 'Error',         color: '#ef4444' },
  };
  const { icon, label, color } = cfg[status] ?? cfg.idle;
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5, color, fontWeight: 700, fontSize: 13 }}>
      {icon} {label}
    </span>
  );
}

function StepCard({ step }) {
  const score10 = (step.score ?? 0) * 10;
  const isFinal = step.action === 'final_answer';
  const border  = isFinal ? '#22c55e' : riskColor(score10);

  return (
    <div style={{
      border: `1.5px solid ${border}`,
      borderRadius: 10,
      padding: '14px 16px',
      marginBottom: 10,
      background: isFinal ? '#f0fdf4' : '#fff',
    }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
        <span style={{ fontWeight: 700, fontSize: 13, color: '#1e293b' }}>
          Step {step.step_number}
        </span>
        {isFinal ? (
          <span style={{ color: '#22c55e', fontWeight: 700, fontSize: 12 }}>✓ Final Answer</span>
        ) : (
          <span style={{
            background: riskColor(score10), color: '#fff',
            borderRadius: 20, padding: '2px 10px', fontSize: 12, fontWeight: 700,
          }}>
            Risk {score10.toFixed(1)}
          </span>
        )}
      </div>

      <Row label="Thought">{step.thought}</Row>
      <Row label="Action">
        <code style={{ background: '#f1f5f9', padding: '1px 6px', borderRadius: 4, fontSize: 12 }}>
          {step.action}
        </code>
        {step.action_input && Object.keys(step.action_input).length > 0 && (
          <span style={{ color: '#6b7280', fontSize: 12 }}>
            {' — '}{JSON.stringify(step.action_input)}
          </span>
        )}
      </Row>
      <div style={{ fontSize: 12, color: '#374151', background: '#f8fafc', borderRadius: 6, padding: '7px 10px', marginTop: 4 }}>
        <strong>Observation: </strong>{step.observation}
      </div>

      {step.failure_types && step.failure_types.length > 0 && (
        <div style={{ marginTop: 8, display: 'flex', gap: 4, flexWrap: 'wrap' }}>
          {step.failure_types.map((f) => (
            <span key={f} style={{
              background: '#fef2f2', color: '#dc2626',
              border: '1px solid #fca5a5', borderRadius: 20,
              padding: '1px 9px', fontSize: 11, fontWeight: 600,
            }}>
              {f}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

function Row({ label, children }) {
  return (
    <div style={{ fontSize: 12, color: '#374151', marginBottom: 4 }}>
      <strong>{label}: </strong>{children}
    </div>
  );
}

function InterventionItem({ intervention }) {
  const colors = { reprompt: '#3b82f6', rollback: '#f59e0b', decompose: '#8b5cf6', halt: '#ef4444' };
  const c = colors[intervention.intervention_type] ?? '#6b7280';
  const preview = intervention.reprompt ?? '';
  return (
    <div style={{ borderLeft: `3px solid ${c}`, paddingLeft: 10, marginBottom: 14 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span style={{ fontWeight: 700, fontSize: 11, color: c, textTransform: 'uppercase', letterSpacing: '0.05em' }}>
          {intervention.intervention_type}
        </span>
        <span style={{ fontSize: 11, color: '#9ca3af' }}>Step {intervention.step_number}</span>
      </div>
      <div style={{ fontSize: 12, color: '#374151', marginTop: 2 }}>{intervention.reason}</div>
      {preview && (
        <div style={{ fontSize: 11, color: '#6b7280', marginTop: 4, fontStyle: 'italic' }}>
          "{preview.length > 130 ? preview.slice(0, 130) + '…' : preview}"
        </div>
      )}
    </div>
  );
}

// ── main dashboard ────────────────────────────────────────────────────────────

export default function App() {
  const [goal, setGoal]       = useState('');
  const [apiKey, setApiKey]   = useState('');
  const [maxSteps, setMaxSteps] = useState(10);

  const { status, steps, interventions, healthScores, finalOutput, error, run, stop } =
    useAgentWebSocket();

  const isRunning = status === 'connecting' || status === 'running';

  const handleRun = () => {
    if (!goal.trim()) return;
    run(goal.trim(), apiKey.trim(), maxSteps);
  };

  return (
    <div style={{ minHeight: '100vh', background: '#f1f5f9', fontFamily: 'system-ui, sans-serif' }}>

      {/* ── header ── */}
      <header style={{
        background: '#0f172a', color: '#fff',
        padding: '0 28px', height: 54,
        display: 'flex', alignItems: 'center', gap: 10,
      }}>
        <Shield size={20} color="#38bdf8" />
        <span style={{ fontWeight: 800, fontSize: 17, letterSpacing: '-0.4px' }}>SentinelAI</span>
        <span style={{ marginLeft: 'auto' }}>
          <StatusBadge status={status} />
        </span>
      </header>

      <div style={{ maxWidth: 1300, margin: '0 auto', padding: '22px 20px' }}>

        {/* ── control panel ── */}
        <div style={{
          background: '#fff', borderRadius: 12, padding: '18px 20px',
          marginBottom: 20, boxShadow: '0 1px 4px rgba(0,0,0,0.07)',
          display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'flex-end',
        }}>
          <Field label="Goal" flex="2 1 300px">
            <input
              value={goal}
              onChange={(e) => setGoal(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && !isRunning && handleRun()}
              placeholder="What should the agent accomplish?"
              disabled={isRunning}
              style={inputStyle}
            />
          </Field>

          <Field label="Gemini API Key" flex="1 1 200px">
            <input
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              type="password"
              placeholder="AIza…"
              disabled={isRunning}
              style={inputStyle}
            />
          </Field>

          <Field label="Max Steps" flex="0 0 90px">
            <input
              value={maxSteps}
              onChange={(e) => setMaxSteps(Number(e.target.value))}
              type="number"
              min={1}
              max={20}
              disabled={isRunning}
              style={inputStyle}
            />
          </Field>

          <div style={{ display: 'flex', gap: 8, paddingBottom: 0 }}>
            <button
              onClick={handleRun}
              disabled={isRunning || !goal.trim()}
              style={{
                ...btnStyle,
                background: isRunning || !goal.trim() ? '#94a3b8' : '#0f172a',
                cursor: isRunning || !goal.trim() ? 'not-allowed' : 'pointer',
              }}
            >
              Run
            </button>
            {isRunning && (
              <button onClick={stop} style={{ ...btnStyle, background: '#ef4444', cursor: 'pointer' }}>
                Stop
              </button>
            )}
          </div>
        </div>

        {/* ── error banner ── */}
        {error && (
          <div style={{
            background: '#fef2f2', border: '1px solid #fca5a5', color: '#dc2626',
            borderRadius: 10, padding: '10px 16px', marginBottom: 16, fontSize: 14,
          }}>
            <strong>Error:</strong> {error}
          </div>
        )}

        {/* ── final answer banner ── */}
        {finalOutput && (
          <div style={{
            background: '#f0fdf4', border: '1.5px solid #86efac',
            borderRadius: 10, padding: '14px 18px', marginBottom: 16,
          }}>
            <div style={{ fontWeight: 700, color: '#15803d', fontSize: 13, marginBottom: 4 }}>
              Final Answer
            </div>
            <div style={{ fontSize: 14, color: '#166534', whiteSpace: 'pre-wrap' }}>
              {finalOutput}
            </div>
          </div>
        )}

        {/* ── main two-column layout ── */}
        <div style={{ display: 'flex', gap: 20, alignItems: 'flex-start' }}>

          {/* left: chart + steps */}
          <div style={{ flex: '1 1 0', minWidth: 0 }}>

            {/* health score chart */}
            <div style={cardStyle}>
              <div style={cardTitle}>Health Score (0 – 10, lower is better)</div>
              <ResponsiveContainer width="100%" height={210}>
                <LineChart data={healthScores} margin={{ top: 6, right: 16, left: -18, bottom: 4 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis
                    dataKey="step"
                    tick={{ fontSize: 11 }}
                    label={{ value: 'Step', position: 'insideBottom', offset: -2, fontSize: 11 }}
                  />
                  <YAxis domain={[0, 10]} tick={{ fontSize: 11 }} />
                  <Tooltip
                    formatter={(v) => [v.toFixed(2), 'Risk Score']}
                    labelFormatter={(l) => `Step ${l}`}
                  />
                  <ReferenceLine
                    y={6}
                    stroke="#ef4444"
                    strokeDasharray="5 3"
                    label={{ value: 'Threshold', fill: '#ef4444', fontSize: 11, position: 'insideTopRight' }}
                  />
                  <Line
                    type="monotone"
                    dataKey="score"
                    stroke="#3b82f6"
                    strokeWidth={2}
                    dot={{ r: 4, fill: '#3b82f6' }}
                    activeDot={{ r: 6 }}
                    isAnimationActive={false}
                  />
                </LineChart>
              </ResponsiveContainer>
              {healthScores.length === 0 && (
                <div style={{ textAlign: 'center', color: '#94a3b8', fontSize: 13, paddingBottom: 8 }}>
                  No data — run the agent to see live health scores.
                </div>
              )}
            </div>

            {/* step cards */}
            <div style={{ marginTop: 4 }}>
              <div style={{ fontWeight: 700, fontSize: 14, color: '#1e293b', marginBottom: 10 }}>
                Steps{' '}
                {steps.length > 0 && (
                  <span style={{ color: '#94a3b8', fontWeight: 400 }}>({steps.length})</span>
                )}
              </div>
              {steps.length === 0 ? (
                <div style={{ ...cardStyle, textAlign: 'center', color: '#94a3b8', fontSize: 13, padding: 36 }}>
                  Agent steps will appear here as they complete.
                </div>
              ) : (
                steps.map((step) => <StepCard key={step.step_number} step={step} />)
              )}
            </div>
          </div>

          {/* right: intervention sidebar */}
          <div style={{ width: 300, flexShrink: 0 }}>
            <div style={{ ...cardStyle, position: 'sticky', top: 16 }}>
              <div style={{ ...cardTitle, display: 'flex', alignItems: 'center', gap: 6, marginBottom: 14 }}>
                <AlertTriangle size={14} color="#f59e0b" />
                Interventions
                {interventions.length > 0 && (
                  <span style={{
                    marginLeft: 'auto', background: '#fef3c7', color: '#d97706',
                    borderRadius: 20, padding: '1px 9px', fontSize: 12, fontWeight: 700,
                  }}>
                    {interventions.length}
                  </span>
                )}
              </div>
              {interventions.length === 0 ? (
                <div style={{ color: '#94a3b8', fontSize: 13, textAlign: 'center', padding: '24px 0' }}>
                  No interventions yet.
                </div>
              ) : (
                interventions.map((iv, i) => <InterventionItem key={i} intervention={iv} />)
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

// ── style constants ───────────────────────────────────────────────────────────

const cardStyle = {
  background: '#fff',
  borderRadius: 12,
  padding: '18px 20px',
  marginBottom: 16,
  boxShadow: '0 1px 4px rgba(0,0,0,0.07)',
};

const cardTitle = {
  fontWeight: 700,
  fontSize: 14,
  color: '#1e293b',
  marginBottom: 12,
};

const inputStyle = {
  width: '100%',
  padding: '8px 11px',
  border: '1.5px solid #e2e8f0',
  borderRadius: 8,
  fontSize: 14,
  outline: 'none',
};

const btnStyle = {
  color: '#fff',
  border: 'none',
  borderRadius: 8,
  padding: '8px 20px',
  fontWeight: 700,
  fontSize: 14,
};

function Field({ label, flex, children }) {
  return (
    <div style={{ flex }}>
      <label style={{ fontSize: 12, fontWeight: 600, color: '#374151', display: 'block', marginBottom: 4 }}>
        {label}
      </label>
      {children}
    </div>
  );
}
