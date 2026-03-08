import React, { useEffect, useCallback } from 'react';
import { Grid, Box, Typography, Alert, AlertTitle } from '@mui/material';
import {
  Computer,
  FolderOpen,
  Topic,
  ShowChart,
} from '@mui/icons-material';
import { StatCard } from '../components/StatCard';
import { DataTable, StatusChip } from '../components/DataTable';
import { useAgentsStore } from '../stores/agentsStore';
import { useTransfersStore } from '../stores/transfersStore';
import { useTopicsStore } from '../stores/topicsStore';
import { useDashboardStore } from '../stores/dashboardStore';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar } from 'recharts';

export const Dashboard: React.FC = () => {
  const { fetchAgents, agents } = useAgentsStore();
  const { fetchActiveTransfers, activeTransfers } = useTransfersStore();
  const { fetchTopics, fetchBufferStats, topics, bufferStats } = useTopicsStore();
  const { updateStats, setLoading, setError } = useDashboardStore();

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      await Promise.all([
        fetchAgents(),
        fetchActiveTransfers(),
        fetchTopics(),
        fetchBufferStats(),
      ]);

      // Calculate stats from fetched data
      updateStats({
        total_agents: agents.length,
        online_agents: agents.filter((a) => a.online).length,
        active_transfers: activeTransfers.filter((t) => t.status === 'sending').length,
        total_topics: topics.length,
        buffer_health: bufferStats ? Math.round((bufferStats.delivered_messages / (bufferStats.total_messages || 1)) * 100) : 0,
        transfers_today: activeTransfers.length,
        total_bytes_sent: activeTransfers.reduce((acc, t) => acc + t.bytes_sent, 0),
      });
      setLoading(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard data');
      setLoading(false);
    }
  }, [fetchAgents, fetchActiveTransfers, fetchTopics, fetchBufferStats, updateStats, setLoading, setError, agents, activeTransfers, topics, bufferStats]);

  useEffect(() => {
    loadData();
    // Refresh data every 30 seconds
    const interval = setInterval(loadData, 30000);
    return () => clearInterval(interval);
  }, [loadData]);

  // Mock chart data - in production, this would come from historical data
  const transferData = [
    { time: '00:00', transfers: 12, bytes: 1024 },
    { time: '04:00', transfers: 8, bytes: 512 },
    { time: '08:00', transfers: 24, bytes: 2048 },
    { time: '12:00', transfers: 36, bytes: 4096 },
    { time: '16:00', transfers: 28, bytes: 3072 },
    { time: '20:00', transfers: 18, bytes: 1536 },
  ];

  const agentStatusData = [
    { name: 'Online', value: agents.filter((a) => a.online).length },
    { name: 'Offline', value: agents.filter((a) => !a.online).length },
  ];

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Dashboard
      </Typography>

      {/* Error Alert */}
      {useDashboardStore.getState().error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError(null)}>
          <AlertTitle>Error</AlertTitle>
          {useDashboardStore.getState().error}
        </Alert>
      )}

      {/* Stats Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={4} lg={2}>
          <StatCard
            title="Total Agents"
            value={agents.length}
            icon={<Computer />}
            color="primary"
            subtitle={`${agents.filter((a) => a.online).length} online`}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={4} lg={2}>
          <StatCard
            title="Active Transfers"
            value={activeTransfers.filter((t) => t.status === 'sending').length}
            icon={<FolderOpen />}
            color="secondary"
            subtitle={`${activeTransfers.length} total`}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={4} lg={2}>
          <StatCard
            title="Topics"
            value={topics.length}
            icon={<Topic />}
            color="success"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={4} lg={2}>
          <StatCard
            title="Buffer Health"
            value={bufferStats ? `${Math.round((bufferStats.delivered_messages / (bufferStats.total_messages || 1)) * 100)}%` : 'N/A'}
            icon={<ShowChart />}
            color="warning"
            subtitle={bufferStats ? `${bufferStats.pending_messages} pending` : 'No data'}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={4} lg={2}>
          <StatCard
            title="Bytes Sent"
            value={activeTransfers.reduce((acc, t) => acc + t.bytes_sent, 0).toLocaleString()}
            icon={<ShowChart />}
            color="primary"
            subtitle="Today"
          />
        </Grid>
      </Grid>

      {/* Charts */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} md={8}>
          <Box sx={{ p: 2, bgcolor: 'background.paper', borderRadius: 2 }}>
            <Typography variant="h6" gutterBottom>
              Transfer Activity (24h)
            </Typography>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={transferData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="time" />
                <YAxis />
                <Tooltip />
                <Line type="monotone" dataKey="transfers" stroke="#90caf9" strokeWidth={2} />
              </LineChart>
            </ResponsiveContainer>
          </Box>
        </Grid>
        <Grid item xs={12} md={4}>
          <Box sx={{ p: 2, bgcolor: 'background.paper', borderRadius: 2 }}>
            <Typography variant="h6" gutterBottom>
              Agent Status
            </Typography>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={agentStatusData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Bar dataKey="value" fill="#4caf50" />
              </BarChart>
            </ResponsiveContainer>
          </Box>
        </Grid>
      </Grid>

      {/* Recent Active Transfers */}
      <Grid container spacing={3}>
        <Grid item xs={12}>
          <DataTable
            title="Active Transfers"
            data={activeTransfers.slice(0, 5) as unknown as Record<string, unknown>[]}
            columns={[
              { key: 'file_name', label: 'File' },
              { key: 'destination', label: 'Destination' },
              {
                key: 'status',
                label: 'Status',
                render: (item: Record<string, unknown>) => (
                  <StatusChip status={String(item.status)} />
                ),
              },
              {
                key: 'bytes_sent',
                label: 'Progress',
                render: (item: Record<string, unknown>) => (
                  <Typography variant="body2">
                    {String(item.percent_complete)}
                  </Typography>
                ),
              },
              {
                key: 'bytes_per_second',
                label: 'Speed',
                render: (item: Record<string, unknown>) => (
                  <Typography variant="body2">
                    {((item.bytes_per_second as number) / 1024).toFixed(1)} KB/s
                  </Typography>
                ),
              },
            ]}
            emptyMessage="No active transfers"
          />
        </Grid>
      </Grid>
    </Box>
  );
};
