import React, { useEffect, useCallback, useState } from 'react';
import {
  Box,
  Typography,
  Grid,
  Card,
  CardContent,
  LinearProgress,
  Chip,
  IconButton,
  Tooltip,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Autocomplete,
  Alert,
  TableContainer,
  Table,
  TableHead,
  TableRow,
  TableBody,
  TableCell,
  Paper,
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  Upload as UploadIcon,
  FolderOpen as FolderOpenIcon,
  History as HistoryIcon,
} from '@mui/icons-material';
import { useTransfersStore } from '../stores/transfersStore';
import { useAgentsStore } from '../stores/agentsStore';
import { tailfsApi } from '../api/client';

export const Transfers: React.FC = () => {
  const {
    activeTransfers,
    history,
    loading,
    fetchActiveTransfers,
    fetchHistory,
  } = useTransfersStore();
  const { agents } = useAgentsStore();
  const [activeTab, setActiveTab] = useState<'active' | 'history'>('active');
  const [sendDialogOpen, setSendDialogOpen] = useState(false);
  const [selectedFile, setSelectedFile] = useState<string>('');
  const [destination, setDestination] = useState<string>('');
  const [sending, setSending] = useState(false);
  const [sendError, setSendError] = useState<string | null>(null);
  const [sendSuccess, setSendSuccess] = useState<string | null>(null);

  const loadData = useCallback(() => {
    fetchActiveTransfers();
    fetchHistory();
  }, [fetchActiveTransfers, fetchHistory]);

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 5000); // Refresh every 5 seconds for active transfers
    return () => clearInterval(interval);
  }, [loadData]);

  const handleSendFile = async () => {
    if (!selectedFile || !destination) return;

    setSending(true);
    setSendError(null);
    setSendSuccess(null);

    try {
      const response = await tailfsApi.sendFile({
        file: selectedFile,
        destination,
        compress: true,
      });
      setSendSuccess(`Transfer initiated: ${response.transfer_id}`);
      setSelectedFile('');
      setDestination('');
      setTimeout(() => {
        setSendDialogOpen(false);
        setSendSuccess(null);
      }, 2000);
      loadData();
    } catch (err) {
      setSendError(err instanceof Error ? err.message : 'Failed to send file');
    } finally {
      setSending(false);
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatSpeed = (bytesPerSecond: number) => {
    return formatBytes(bytesPerSecond) + '/s';
  };

  const formatETA = (seconds: number) => {
    if (seconds < 60) return `${Math.round(seconds)}s`;
    if (seconds < 3600) return `${Math.round(seconds / 60)}m`;
    return `${Math.round(seconds / 3600)}h ${Math.round((seconds % 3600) / 60)}m`;
  };

  const receiveAgents = agents.filter((agent) => agent.online);

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">Transfers</Typography>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Button
            variant={activeTab === 'active' ? 'contained' : 'outlined'}
            startIcon={<FolderOpenIcon />}
            onClick={() => setActiveTab('active')}
          >
            Active
          </Button>
          <Button
            variant={activeTab === 'history' ? 'contained' : 'outlined'}
            startIcon={<HistoryIcon />}
            onClick={() => setActiveTab('history')}
          >
            History
          </Button>
          <Tooltip title="New Transfer">
            <IconButton color="primary" onClick={() => setSendDialogOpen(true)}>
              <UploadIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Refresh">
            <IconButton onClick={loadData} disabled={loading}>
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>

      {/* Active Transfers */}
      {activeTab === 'active' && (
        <Grid container spacing={3}>
          {activeTransfers.map((transfer) => (
            <Grid item xs={12} key={transfer.transfer_id}>
              <Card>
                <CardContent>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                      <FolderOpenIcon color="primary" sx={{ fontSize: 32 }} />
                      <Box>
                        <Typography variant="h6">{transfer.file_name}</Typography>
                        <Typography variant="body2" color="text.secondary">
                          {transfer.source} → {transfer.destination}
                        </Typography>
                      </Box>
                    </Box>
                    <Chip
                      label={transfer.status}
                      color={
                        transfer.status === 'completed'
                          ? 'success'
                          : transfer.status === 'sending'
                          ? 'info'
                          : transfer.status === 'failed'
                          ? 'error'
                          : 'warning'
                      }
                    />
                  </Box>

                  <LinearProgress
                    variant="determinate"
                    value={parseFloat(transfer.percent_complete)}
                    sx={{ height: 8, borderRadius: 4, mb: 2 }}
                  />

                  <Grid container spacing={2}>
                    <Grid item xs={6} sm={3}>
                      <Typography variant="body2" color="text.secondary">
                        Progress
                      </Typography>
                      <Typography variant="body1">
                        {transfer.percent_complete}
                      </Typography>
                    </Grid>
                    <Grid item xs={6} sm={3}>
                      <Typography variant="body2" color="text.secondary">
                        Transferred
                      </Typography>
                      <Typography variant="body1">
                        {formatBytes(transfer.bytes_sent)} / {formatBytes(transfer.bytes_total)}
                      </Typography>
                    </Grid>
                    <Grid item xs={6} sm={3}>
                      <Typography variant="body2" color="text.secondary">
                        Speed
                      </Typography>
                      <Typography variant="body1">
                        {formatSpeed(transfer.bytes_per_second)}
                      </Typography>
                    </Grid>
                    <Grid item xs={6} sm={3}>
                      <Typography variant="body2" color="text.secondary">
                        ETA
                      </Typography>
                      <Typography variant="body1">
                        {transfer.eta_seconds > 0 ? formatETA(transfer.eta_seconds) : 'Complete'}
                      </Typography>
                    </Grid>
                  </Grid>
                </CardContent>
              </Card>
            </Grid>
          ))}

          {activeTransfers.length === 0 && (
            <Box sx={{ textAlign: 'center', py: 8 }}>
              <FolderOpenIcon sx={{ fontSize: 64, color: 'text.secondary', mb: 2 }} />
              <Typography variant="h6" color="text.secondary">
                No active transfers
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Start a new transfer to see progress here
              </Typography>
            </Box>
          )}
        </Grid>
      )}

      {/* Transfer History */}
      {activeTab === 'history' && (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>File</TableCell>
                <TableCell>Source</TableCell>
                <TableCell>Destination</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Size</TableCell>
                <TableCell>Completed</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {history.slice(0, 50).map((transfer) => (
                <TableRow key={transfer.transfer_id}>
                  <TableCell>{transfer.file_name}</TableCell>
                  <TableCell>{transfer.source}</TableCell>
                  <TableCell>{transfer.destination}</TableCell>
                  <TableCell>
                    <Chip
                      label={transfer.status}
                      size="small"
                      color={
                        transfer.status === 'completed'
                          ? 'success'
                          : transfer.status === 'failed'
                          ? 'error'
                          : 'default'
                      }
                    />
                  </TableCell>
                  <TableCell>{formatBytes(transfer.file_size)}</TableCell>
                  <TableCell>
                    {transfer.completed_at
                      ? new Date(transfer.completed_at).toLocaleString()
                      : '-'}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {/* Send File Dialog */}
      <Dialog open={sendDialogOpen} onClose={() => setSendDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Send File</DialogTitle>
        <DialogContent>
          {sendError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {sendError}
            </Alert>
          )}
          {sendSuccess && (
            <Alert severity="success" sx={{ mb: 2 }}>
              {sendSuccess}
            </Alert>
          )}
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: 1 }}>
            <TextField
              fullWidth
              label="File Path"
              placeholder="/path/to/file"
              value={selectedFile}
              onChange={(e) => setSelectedFile(e.target.value)}
              required
            />
            <Autocomplete
              fullWidth
              options={receiveAgents.map((a) => a.name)}
              renderInput={(params) => (
                <TextField {...params} label="Destination Agent" required />
              )}
              value={destination}
              onChange={(_, newValue) => setDestination(newValue || '')}
              disabled={sending}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSendDialogOpen(false)} disabled={sending}>
            Cancel
          </Button>
          <Button
            onClick={handleSendFile}
            variant="contained"
            disabled={!selectedFile || !destination || sending}
          >
            {sending ? 'Sending...' : 'Send'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};
