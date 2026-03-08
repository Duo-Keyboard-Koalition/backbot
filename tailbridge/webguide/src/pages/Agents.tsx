import React, { useEffect, useCallback, useState } from 'react';
import {
  Box,
  Typography,
  Grid,
  Card,
  CardContent,
  Chip,
  IconButton,
  TextField,
  InputAdornment,
  FormControlLabel,
  Switch,
  Tooltip,
  Divider,
  Avatar,
} from '@mui/material';
import {
  Search as SearchIcon,
  Refresh as RefreshIcon,
  CheckCircle as CheckCircleIcon,
  Cancel as CancelIcon,
  Dns as DnsIcon,
} from '@mui/icons-material';
import { useAgentsStore } from '../stores/agentsStore';

export const Agents: React.FC = () => {
  const { agents, loading, error, lastUpdated, fetchAgents } = useAgentsStore();
  const [searchTerm, setSearchTerm] = useState('');
  const [showOnlineOnly, setShowOnlineOnly] = useState(false);

  const loadData = useCallback(() => {
    fetchAgents();
  }, [fetchAgents]);

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 30000);
    return () => clearInterval(interval);
  }, [loadData]);

  const filteredAgents = agents.filter((agent) => {
    const matchesSearch =
      agent.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      agent.hostname.toLowerCase().includes(searchTerm.toLowerCase()) ||
      agent.ip.includes(searchTerm);
    const matchesOnline = showOnlineOnly ? agent.online : true;
    return matchesSearch && matchesOnline;
  });

  const onlineCount = agents.filter((a) => a.online).length;
  const offlineCount = agents.length - onlineCount;

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">Agents</Typography>
        <Tooltip title="Refresh">
          <IconButton onClick={loadData} disabled={loading}>
            <RefreshIcon />
          </IconButton>
        </Tooltip>
      </Box>

      {/* Filters */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Grid container spacing={2} alignItems="center">
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                placeholder="Search by name, hostname, or IP..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <SearchIcon />
                    </InputAdornment>
                  ),
                }}
                size="small"
              />
            </Grid>
            <Grid item xs={12} md={3}>
              <FormControlLabel
                control={
                  <Switch
                    checked={showOnlineOnly}
                    onChange={(e) => setShowOnlineOnly(e.target.checked)}
                    color="success"
                  />
                }
                label="Online Only"
              />
            </Grid>
            <Grid item xs={12} md={3}>
              <Box sx={{ display: 'flex', gap: 1, justifyContent: 'flex-end' }}>
                <Chip
                  label={`${onlineCount} online`}
                  color="success"
                  size="small"
                  icon={<CheckCircleIcon />}
                />
                <Chip
                  label={`${offlineCount} offline`}
                  color="error"
                  size="small"
                  icon={<CancelIcon />}
                />
              </Box>
            </Grid>
          </Grid>
        </CardContent>
      </Card>

      {/* Error Alert */}
      {error && (
        <Box sx={{ mb: 3 }}>
          <Typography color="error">{error}</Typography>
        </Box>
      )}

      {/* Agent Cards */}
      <Grid container spacing={3}>
        {filteredAgents.map((agent) => (
          <Grid item xs={12} md={6} lg={4} key={agent.name}>
            <Card
              sx={{
                height: '100%',
                borderLeft: agent.online ? '4px solid #4caf50' : '4px solid #f44336',
                cursor: 'pointer',
                transition: 'transform 0.2s, box-shadow 0.2s',
                '&:hover': {
                  transform: 'translateY(-4px)',
                  boxShadow: 8,
                },
              }}
            >
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <Avatar
                      sx={{
                        bgcolor: agent.online ? 'success.main' : 'error.main',
                        width: 48,
                        height: 48,
                      }}
                    >
                      <DnsIcon />
                    </Avatar>
                    <Box>
                      <Typography variant="h6">{agent.name}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        {agent.hostname}
                      </Typography>
                    </Box>
                  </Box>
                  <Chip
                    label={agent.online ? 'Online' : 'Offline'}
                    color={agent.online ? 'success' : 'error'}
                    size="small"
                  />
                </Box>

                <Divider sx={{ my: 2 }} />

                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                  <Typography variant="body2" color="text.secondary">
                    <strong>IP:</strong> {agent.ip}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    <strong>Last Seen:</strong>{' '}
                    {new Date(agent.last_seen).toLocaleString()}
                  </Typography>
                </Box>

                {/* Gateways */}
                {agent.gateways && agent.gateways.length > 0 && (
                  <Box sx={{ mt: 2 }}>
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      <strong>Services:</strong>
                    </Typography>
                    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                      {agent.gateways.map((gateway, idx) => (
                        <Chip
                          key={idx}
                          label={`${gateway.service}:${gateway.port}`}
                          size="small"
                          variant="outlined"
                        />
                      ))}
                    </Box>
                  </Box>
                )}
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      {filteredAgents.length === 0 && !loading && (
        <Box sx={{ textAlign: 'center', py: 8 }}>
          <Typography variant="h6" color="text.secondary">
            No agents found
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Try adjusting your search or filters
          </Typography>
        </Box>
      )}

      {lastUpdated && (
        <Box sx={{ mt: 3, textAlign: 'center' }}>
          <Typography variant="caption" color="text.secondary">
            Last updated: {lastUpdated.toLocaleString()}
          </Typography>
        </Box>
      )}
    </Box>
  );
};
