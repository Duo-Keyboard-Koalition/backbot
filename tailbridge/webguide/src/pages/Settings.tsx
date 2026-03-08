import React from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  TextField,
  Button,
  Divider,
  Alert,
  Link,
} from '@mui/material';
import { Save as SaveIcon, Info as InfoIcon } from '@mui/icons-material';

export const Settings: React.FC = () => {
  const [taila2aUrl, setTaila2aUrl] = React.useState('http://localhost:8080');
  const [tailfsUrl, setTailfsUrl] = React.useState('http://localhost:8081');
  const [refreshInterval, setRefreshInterval] = React.useState('30');
  const [saved, setSaved] = React.useState(false);

  const handleSave = () => {
    // In production, save to localStorage or backend
    localStorage.setItem('taila2aUrl', taila2aUrl);
    localStorage.setItem('tailfsUrl', tailfsUrl);
    localStorage.setItem('refreshInterval', refreshInterval);
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Settings
      </Typography>

      {saved && (
        <Alert severity="success" sx={{ mb: 3 }}>
          Settings saved successfully!
        </Alert>
      )}

      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                API Endpoints
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <TextField
                fullWidth
                label="Taila2a API URL"
                value={taila2aUrl}
                onChange={(e) => setTaila2aUrl(e.target.value)}
                margin="normal"
                helperText="Agent-to-Agent communication API"
              />
              <TextField
                fullWidth
                label="TailFS API URL"
                value={tailfsUrl}
                onChange={(e) => setTailfsUrl(e.target.value)}
                margin="normal"
                helperText="File transfer API"
              />
            </CardContent>
          </Card>

          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Display Settings
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <TextField
                fullWidth
                label="Refresh Interval (seconds)"
                type="number"
                value={refreshInterval}
                onChange={(e) => setRefreshInterval(e.target.value)}
                margin="normal"
                helperText="How often to refresh data from APIs"
              />
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                About Webguide
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
                <InfoIcon color="primary" />
                <Typography variant="body1">
                  Webguide is the web-based dashboard for Tailbridge
                </Typography>
              </Box>
              <Typography variant="body2" color="text.secondary" paragraph>
                Version: 1.0.0
              </Typography>
              <Typography variant="body2" color="text.secondary" paragraph>
                Webguide provides real-time monitoring and management for your
                Tailbridge ecosystem, including Taila2a agent-to-agent communication
                and TailFS secure file transfers.
              </Typography>
            </CardContent>
          </Card>

          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Quick Links
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                <Link href="https://github.com/tailbridge" target="_blank" rel="noopener">
                  GitHub Repository
                </Link>
                <Link href="https://tailscale.com" target="_blank" rel="noopener">
                  Tailscale Documentation
                </Link>
                <Link href="/eng_nbk/README.md" target="_blank" rel="noopener">
                  Engineering Notebook
                </Link>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Box sx={{ mt: 3, display: 'flex', justifyContent: 'flex-end' }}>
        <Button
          variant="contained"
          startIcon={<SaveIcon />}
          onClick={handleSave}
          size="large"
        >
          Save Settings
        </Button>
      </Box>
    </Box>
  );
};
