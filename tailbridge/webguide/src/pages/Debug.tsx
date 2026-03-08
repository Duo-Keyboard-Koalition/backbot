import React, { useEffect, useState } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Button,
  Alert,
  AlertTitle,
  IconButton,
  Tooltip,
  TextField,
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  BugReport as BugIcon,
  Dns as DnsIcon,
  NetworkCheck as NetworkCheckIcon,
  Key as KeyIcon,
  Info as InfoIcon,
  PlayArrow as PlayArrowIcon,
  Clear as ClearIcon,
} from '@mui/icons-material';
import {
  getConnectionInfo,
  getNetworkStatus,
  testApiConnectivity,
  logSystemInfo,
  isAuthKeyConfigured,
  getMaskedAuthKey,
  type ConnectionInfo,
  type NetworkStatus,
} from '../utils/debug';
import { testTailscaleConnection } from '../api/tailscale';

interface ApiTestResult {
  name: string;
  url: string;
  success: boolean;
  status?: number;
  error?: string;
  responseTime?: number;
}

interface LogEntry {
  timestamp: string;
  level: 'info' | 'warn' | 'error' | 'log';
  message: string;
}

export const Debug: React.FC = () => {
  const [connectionInfo, setConnectionInfo] = useState<ConnectionInfo | null>(null);
  const [networkStatus, setNetworkStatus] = useState<NetworkStatus | null>(null);
  const [apiTests, setApiTests] = useState<ApiTestResult[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [testing, setTesting] = useState(false);
  const [customUrl, setCustomUrl] = useState('http://localhost:8080');
  const [tailscaleTestResult, setTailscaleTestResult] = useState<{
    success: boolean;
    devices?: Array<{ id: string; name: string; addresses: string[]; online: boolean }>;
    error?: string;
    responseTime?: number;
  } | null>(null);
  const [tailscaleTesting, setTailscaleTesting] = useState(false);

  useEffect(() => {
    // Load initial info
    setConnectionInfo(getConnectionInfo());
    setNetworkStatus(getNetworkStatus());
    
    // Log system info
    logSystemInfo();
    
    // Add initial log entry
    addLog('info', 'Debug page loaded');
    
    // Listen for online/offline events
    const handleOnline = () => {
      setNetworkStatus(getNetworkStatus());
      addLog('info', 'Network status: Online');
    };
    
    const handleOffline = () => {
      setNetworkStatus(getNetworkStatus());
      addLog('warn', 'Network status: Offline');
    };
    
    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);
    
    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  const addLog = (level: 'info' | 'warn' | 'error' | 'log', message: string) => {
    setLogs((prev) => [
      ...prev,
      {
        timestamp: new Date().toISOString(),
        level,
        message,
      },
    ]);
  };

  const runApiTests = async () => {
    setTesting(true);
    addLog('info', 'Starting API connectivity tests...');

    const tests = [
      { name: 'Taila2a API', url: `${getConnectionInfo().taila2aUrl}/agents` },
      { name: 'TailFS API', url: `${getConnectionInfo().tailfsUrl}/transfers` },
      { name: 'Custom Endpoint', url: customUrl },
    ];

    const results: ApiTestResult[] = [];

    for (const test of tests) {
      addLog('info', `Testing ${test.name}: ${test.url}`);
      const result = await testApiConnectivity(test.url);

      results.push({
        name: test.name,
        url: test.url,
        ...result,
      });

      if (result.success) {
        addLog('info', `${test.name} - Success (${result.status}) in ${result.responseTime?.toFixed(0)}ms`);
      } else {
        addLog('error', `${test.name} - Failed: ${result.error}`);
      }
    }

    setApiTests(results);
    setTesting(false);
    addLog('info', 'API tests completed');
  };

  const testTailscaleRealApi = async () => {
    setTailscaleTesting(true);
    addLog('info', 'Testing REAL Tailscale API connection...');
    addLog('info', `Auth key: ${getMaskedAuthKey()}`);

    try {
      const result = await testTailscaleConnection();
      setTailscaleTestResult(result);

      if (result.success) {
        addLog('info', `✅ Tailscale API SUCCESS! Found ${result.devices?.length || 0} devices in ${result.responseTime?.toFixed(0)}ms`);
        result.devices?.forEach(device => {
          addLog('info', `  Device: ${device.name} (${device.addresses.join(', ')}) - ${device.online ? 'Online' : 'Offline'}`);
        });
      } else {
        addLog('error', `❌ Tailscale API FAILED: ${result.error}`);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      addLog('error', `❌ Tailscale API ERROR: ${errorMessage}`);
      setTailscaleTestResult({
        success: false,
        error: errorMessage,
      });
    }

    setTailscaleTesting(false);
  };

  const clearLogs = () => {
    setLogs([]);
    addLog('info', 'Logs cleared');
  };

  const exportLogs = () => {
    const logText = logs.map((log) => `[${log.timestamp}] [${log.level.toUpperCase()}] ${log.message}`).join('\n');
    const blob = new Blob([logText], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `webguide-debug-${new Date().toISOString().split('T')[0]}.log`;
    a.click();
    URL.revokeObjectURL(url);
    addLog('info', 'Logs exported');
  };

  if (!connectionInfo) return null;

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">Debug & Diagnostics</Typography>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Tooltip title="Refresh">
            <IconButton onClick={() => {
              setConnectionInfo(getConnectionInfo());
              setNetworkStatus(getNetworkStatus());
              addLog('info', 'Debug info refreshed');
            }}>
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>

      {/* Auth Key Status */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
            <KeyIcon color={isAuthKeyConfigured() ? 'success' : 'error'} />
            <Typography variant="h6">Tailscale Authentication</Typography>
          </Box>
          <Grid container spacing={2}>
            <Grid item xs={12} md={6}>
              <Alert
                severity={isAuthKeyConfigured() ? 'success' : 'error'}
                icon={<KeyIcon />}
              >
                <AlertTitle>Auth Key Status</AlertTitle>
                {isAuthKeyConfigured() ? (
                  <Box sx={{ mt: 1 }}>
                    <Typography variant="body2">
                      ✅ Auth key is configured
                    </Typography>
                    <Typography variant="body2" sx={{ mt: 0.5, fontFamily: 'monospace' }}>
                      {getMaskedAuthKey()}
                    </Typography>
                  </Box>
                ) : (
                  <Box sx={{ mt: 1 }}>
                    <Typography variant="body2">
                      ❌ Auth key is missing or invalid
                    </Typography>
                    <Typography variant="body2" sx={{ mt: 0.5 }}>
                      Set VITE_TAILSCALE_AUTH_KEY in .env file
                    </Typography>
                  </Box>
                )}
              </Alert>
            </Grid>
            <Grid item xs={12} md={6}>
              <Box sx={{ p: 2, bgcolor: 'background.default', borderRadius: 1 }}>
                <Typography variant="body2" color="text.secondary" gutterBottom>
                  Expected format:
                </Typography>
                <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                  tskey-auth-XXXXXXXXXX...
                </Typography>
                <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                  The auth key is used to authenticate with Tailscale when establishing connections.
                </Typography>
              </Box>
            </Grid>
          </Grid>
        </CardContent>
      </Card>

      {/* Connection Info */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
            <DnsIcon color="primary" />
            <Typography variant="h6">Connection Information</Typography>
          </Box>
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Property</TableCell>
                  <TableCell>Value</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                <TableRow>
                  <TableCell>Debug Mode</TableCell>
                  <TableCell>
                    <Chip
                      label={connectionInfo.debugEnabled ? 'Enabled' : 'Disabled'}
                      color={connectionInfo.debugEnabled ? 'success' : 'default'}
                      size="small"
                    />
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell>Taila2a URL</TableCell>
                  <TableCell sx={{ fontFamily: 'monospace' }}>
                    {connectionInfo.taila2aUrl}
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell>TailFS URL</TableCell>
                  <TableCell sx={{ fontFamily: 'monospace' }}>
                    {connectionInfo.tailfsUrl}
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell>Refresh Interval</TableCell>
                  <TableCell>{connectionInfo.refreshInterval / 1000}s</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell>Network Status</TableCell>
                  <TableCell>
                    <Chip
                      label={networkStatus?.online ? 'Online' : 'Offline'}
                      color={networkStatus?.online ? 'success' : 'error'}
                      size="small"
                      icon={<NetworkCheckIcon />}
                    />
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell>Network Type</TableCell>
                  <TableCell>{networkStatus?.effectiveType || 'Unknown'}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell>Downlink</TableCell>
                  <TableCell>{networkStatus?.downlink ? `${networkStatus.downlink} Mbps` : 'Unknown'}</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>
        </CardContent>
      </Card>

      {/* Real Tailscale API Test */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
            <DnsIcon color="success" />
            <Typography variant="h6">REAL Tailscale API Test</Typography>
          </Box>
          <Alert severity="info" sx={{ mb: 2 }}>
            <AlertTitle>Live Tailscale API Connection</AlertTitle>
            This test makes a <strong>real API call</strong> to Tailscale's management API at{' '}
            <code>https://api.tailscale.com/api/v2</code> using your auth key.
            No mocks - this tests the actual connection to your tailnet!
          </Alert>
          <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
            <Button
              variant="contained"
              color="success"
              startIcon={<PlayArrowIcon />}
              onClick={testTailscaleRealApi}
              disabled={tailscaleTesting || !isAuthKeyConfigured()}
              fullWidth
            >
              {tailscaleTesting ? 'Testing...' : 'Test REAL Tailscale API Connection'}
            </Button>
          </Box>
          {tailscaleTestResult && (
            <Box>
              <Alert
                severity={tailscaleTestResult.success ? 'success' : 'error'}
                sx={{ mb: 2 }}
              >
                <AlertTitle>
                  {tailscaleTestResult.success ? '✅ SUCCESS' : '❌ FAILED'}
                </AlertTitle>
                {tailscaleTestResult.success ? (
                  <Box>
                    <Typography variant="body2">
                      Response time: {tailscaleTestResult.responseTime?.toFixed(0)}ms
                    </Typography>
                    <Typography variant="body2">
                      Devices found: {tailscaleTestResult.devices?.length || 0}
                    </Typography>
                  </Box>
                ) : (
                  <Typography variant="body2">
                    Error: {tailscaleTestResult.error}
                  </Typography>
                )}
              </Alert>
              {tailscaleTestResult.success && tailscaleTestResult.devices && (
                <TableContainer>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Device Name</TableCell>
                        <TableCell>IP Addresses</TableCell>
                        <TableCell>Status</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {tailscaleTestResult.devices.map((device) => (
                        <TableRow key={device.id}>
                          <TableCell sx={{ fontFamily: 'monospace' }}>{device.name}</TableCell>
                          <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.85em' }}>
                            {device.addresses.join(', ')}
                          </TableCell>
                          <TableCell>
                            <Chip
                              label={device.online ? 'Online' : 'Offline'}
                              color={device.online ? 'success' : 'default'}
                              size="small"
                            />
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              )}
            </Box>
          )}
        </CardContent>
      </Card>

      {/* API Connectivity Tests */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
            <BugIcon color="secondary" />
            <Typography variant="h6">API Connectivity Tests</Typography>
          </Box>
          <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
            <TextField
              size="small"
              label="Custom URL to test"
              value={customUrl}
              onChange={(e) => setCustomUrl(e.target.value)}
              sx={{ flexGrow: 1 }}
            />
            <Button
              variant="contained"
              startIcon={<PlayArrowIcon />}
              onClick={runApiTests}
              disabled={testing}
            >
              {testing ? 'Testing...' : 'Run Tests'}
            </Button>
          </Box>
          {apiTests.length > 0 && (
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>API</TableCell>
                    <TableCell>URL</TableCell>
                    <TableCell>Status</TableCell>
                    <TableCell>Response Time</TableCell>
                    <TableCell>Error</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {apiTests.map((test, idx) => (
                    <TableRow key={idx}>
                      <TableCell>{test.name}</TableCell>
                      <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.85em' }}>
                        {test.url}
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={test.success ? `OK (${test.status})` : 'Failed'}
                          color={test.success ? 'success' : 'error'}
                          size="small"
                        />
                      </TableCell>
                      <TableCell>
                        {test.responseTime ? `${test.responseTime.toFixed(0)}ms` : '-'}
                      </TableCell>
                      <TableCell>{test.error || '-'}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </CardContent>
      </Card>

      {/* Debug Logs */}
      <Card>
        <CardContent>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <InfoIcon color="info" />
              <Typography variant="h6">Debug Logs</Typography>
            </Box>
            <Box sx={{ display: 'flex', gap: 1 }}>
              <Button
                size="small"
                startIcon={<ClearIcon />}
                onClick={clearLogs}
              >
                Clear
              </Button>
              <Button
                size="small"
                startIcon={<BugIcon />}
                onClick={exportLogs}
              >
                Export
              </Button>
            </Box>
          </Box>
          <Box
            sx={{
              height: 300,
              overflow: 'auto',
              bgcolor: 'background.default',
              borderRadius: 1,
              p: 2,
              fontFamily: 'monospace',
              fontSize: '0.85em',
            }}
          >
            {logs.length === 0 ? (
              <Typography color="text.secondary">No logs yet</Typography>
            ) : (
              logs.map((log, idx) => (
                <Box
                  key={idx}
                  sx={{
                    mb: 0.5,
                    color:
                      log.level === 'error'
                        ? 'error.main'
                        : log.level === 'warn'
                        ? 'warning.main'
                        : log.level === 'info'
                        ? 'info.main'
                        : 'text.primary',
                  }}
                >
                  <Typography component="span" sx={{ opacity: 0.7 }}>
                    [{new Date(log.timestamp).toLocaleTimeString()}]
                  </Typography>{' '}
                  <Typography component="span" sx={{ fontWeight: 'bold' }}>
                    [{log.level.toUpperCase()}]
                  </Typography>{' '}
                  {log.message}
                </Box>
              ))
            )}
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
};
