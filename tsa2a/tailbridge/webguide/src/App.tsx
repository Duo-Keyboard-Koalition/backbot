import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Layout } from './components/Layout';
import { Dashboard } from './pages/Dashboard';
import { Agents } from './pages/Agents';
import { Transfers } from './pages/Transfers';
import { Topics } from './pages/Topics';
import { Settings } from './pages/Settings';
import { Debug } from './pages/Debug';

export const App: React.FC = () => {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/agents" element={<Agents />} />
          <Route path="/transfers" element={<Transfers />} />
          <Route path="/topics" element={<Topics />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="/debug" element={<Debug />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Layout>
    </Router>
  );
};
