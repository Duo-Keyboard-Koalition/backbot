import React, { useEffect, useCallback } from 'react';
import {
  Box,
  Typography,
  Grid,
  Card,
  CardContent,
  Chip,
  IconButton,
  Tooltip,
  TableContainer,
  Table,
  TableHead,
  TableRow,
  TableBody,
  TableCell,
  Paper,
  Avatar,
  AvatarGroup,
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  Topic as TopicIcon,
  People as PeopleIcon,
  Storage as StorageIcon,
} from '@mui/icons-material';
import { useTopicsStore } from '../stores/topicsStore';
import { StatusChip } from '../components/DataTable';

export const Topics: React.FC = () => {
  const {
    topics,
    consumers,
    bufferStats,
    loading,
    fetchTopics,
    fetchConsumers,
    fetchBufferStats,
  } = useTopicsStore();

  const loadData = useCallback(() => {
    fetchTopics();
    fetchConsumers();
    fetchBufferStats();
  }, [fetchTopics, fetchConsumers, fetchBufferStats]);

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 30000);
    return () => clearInterval(interval);
  }, [loadData]);

  const totalMessages = topics.reduce((acc, topic) => acc + topic.message_count, 0);
  const totalConsumers = consumers.length;

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">Topics & Consumers</Typography>
        <Tooltip title="Refresh">
          <IconButton onClick={loadData} disabled={loading}>
            <RefreshIcon />
          </IconButton>
        </Tooltip>
      </Box>

      {/* Stats Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={4}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <Avatar sx={{ bgcolor: 'primary.main' }}>
                  <TopicIcon />
                </Avatar>
                <Box>
                  <Typography variant="body2" color="text.secondary">
                    Total Topics
                  </Typography>
                  <Typography variant="h4">{topics.length}</Typography>
                </Box>
              </Box>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={4}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <Avatar sx={{ bgcolor: 'secondary.main' }}>
                  <PeopleIcon />
                </Avatar>
                <Box>
                  <Typography variant="body2" color="text.secondary">
                    Total Consumers
                  </Typography>
                  <Typography variant="h4">{totalConsumers}</Typography>
                </Box>
              </Box>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={4}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <Avatar sx={{ bgcolor: 'success.main' }}>
                  <StorageIcon />
                </Avatar>
                <Box>
                  <Typography variant="body2" color="text.secondary">
                    Total Messages
                  </Typography>
                  <Typography variant="h4">{totalMessages.toLocaleString()}</Typography>
                </Box>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Buffer Stats */}
      {bufferStats && (
        <Card sx={{ mb: 4 }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Message Buffer Statistics
            </Typography>
            <Grid container spacing={3}>
              <Grid item xs={6} sm={3}>
                <Typography variant="body2" color="text.secondary">
                  Total Messages
                </Typography>
                <Typography variant="h5">{bufferStats.total_messages}</Typography>
              </Grid>
              <Grid item xs={6} sm={3}>
                <Typography variant="body2" color="text.secondary">
                  Pending
                </Typography>
                <Typography variant="h5" color="warning.main">
                  {bufferStats.pending_messages}
                </Typography>
              </Grid>
              <Grid item xs={6} sm={3}>
                <Typography variant="body2" color="text.secondary">
                  Failed
                </Typography>
                <Typography variant="h5" color="error.main">
                  {bufferStats.failed_messages}
                </Typography>
              </Grid>
              <Grid item xs={6} sm={3}>
                <Typography variant="body2" color="text.secondary">
                  Delivered
                </Typography>
                <Typography variant="h5" color="success.main">
                  {bufferStats.delivered_messages}
                </Typography>
              </Grid>
            </Grid>
            {bufferStats.oldest_message_age > 0 && (
              <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                Oldest message age: {Math.round(bufferStats.oldest_message_age / 60)} minutes
              </Typography>
            )}
          </CardContent>
        </Card>
      )}

      {/* Topics Table */}
      <Grid container spacing={3}>
        <Grid item xs={12} lg={6}>
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Topic</TableCell>
                  <TableCell align="right">Messages</TableCell>
                  <TableCell>Consumers</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {topics.map((topic) => (
                  <TableRow key={topic.name}>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <TopicIcon color="primary" fontSize="small" />
                        <Typography variant="body2">{topic.name}</Typography>
                      </Box>
                    </TableCell>
                    <TableCell align="right">
                      <Chip
                        label={topic.message_count.toLocaleString()}
                        size="small"
                        color="primary"
                        variant="outlined"
                      />
                    </TableCell>
                    <TableCell>
                      <AvatarGroup max={3}>
                        {topic.consumers.map((consumer, idx) => (
                          <Avatar key={idx}>{consumer[0].toUpperCase()}</Avatar>
                        ))}
                      </AvatarGroup>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </Grid>

        {/* Consumers Table */}
        <Grid item xs={12} lg={6}>
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Consumer</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Topics</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {consumers.map((consumer) => (
                  <TableRow key={consumer.name}>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <PeopleIcon color="secondary" fontSize="small" />
                        <Typography variant="body2">{consumer.name}</Typography>
                      </Box>
                    </TableCell>
                    <TableCell>
                      <StatusChip status={consumer.status} />
                    </TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                        {consumer.topics.map((topic, idx) => (
                          <Chip
                            key={idx}
                            label={topic}
                            size="small"
                            variant="outlined"
                          />
                        ))}
                      </Box>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </Grid>
      </Grid>

      {topics.length === 0 && consumers.length === 0 && !loading && (
        <Box sx={{ textAlign: 'center', py: 8 }}>
          <TopicIcon sx={{ fontSize: 64, color: 'text.secondary', mb: 2 }} />
          <Typography variant="h6" color="text.secondary">
            No topics or consumers found
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Topics and consumers will appear here when active
          </Typography>
        </Box>
      )}
    </Box>
  );
};
