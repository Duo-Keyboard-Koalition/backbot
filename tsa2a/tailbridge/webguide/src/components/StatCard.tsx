import React from 'react';
import { Card, CardContent, Typography, Box, LinearProgress, Chip } from '@mui/material';
import { styled } from '@mui/material/styles';

interface StatCardProps {
  title: string;
  value: string | number;
  icon?: React.ReactNode;
  trend?: 'up' | 'down' | 'neutral';
  trendValue?: string;
  subtitle?: string;
  color?: 'primary' | 'secondary' | 'success' | 'warning' | 'error';
}

const StatCardRoot = styled(Card)(({ theme }) => ({
  height: '100%',
  transition: 'transform 0.2s, box-shadow 0.2s',
  '&:hover': {
    transform: 'translateY(-4px)',
    boxShadow: theme.shadows[8],
  },
}));

export const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  icon,
  trend,
  trendValue,
  subtitle,
  color = 'primary',
}) => {
  const colorMap = {
    primary: 'primary.main',
    secondary: 'secondary.main',
    success: 'success.main',
    warning: 'warning.main',
    error: 'error.main',
  };

  return (
    <StatCardRoot>
      <CardContent>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <Box>
            <Typography color="text.secondary" variant="body2" gutterBottom>
              {title}
            </Typography>
            <Typography variant="h4" component="div" gutterBottom>
              {value}
            </Typography>
            {subtitle && (
              <Typography variant="body2" color="text.secondary">
                {subtitle}
              </Typography>
            )}
          </Box>
          {icon && (
            <Box
              sx={{
                bgcolor: `${colorMap[color]}.20`,
                color: `${colorMap[color]}`,
                borderRadius: 2,
                p: 1,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              {icon}
            </Box>
          )}
        </Box>
        {trend && trendValue && (
          <Box sx={{ mt: 2, display: 'flex', alignItems: 'center', gap: 1 }}>
            <Chip
              label={`${trend === 'up' ? '↑' : trend === 'down' ? '↓' : '→'} ${trendValue}`}
              size="small"
              color={trend === 'up' ? 'success' : trend === 'down' ? 'error' : 'default'}
            />
          </Box>
        )}
      </CardContent>
    </StatCardRoot>
  );
};

interface ProgressCardProps {
  title: string;
  value: number;
  total: number;
  unit?: string;
}

export const ProgressCard: React.FC<ProgressCardProps> = ({ title, value, total, unit = '' }) => {
  const percentage = total > 0 ? Math.round((value / total) * 100) : 0;

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          {title}
        </Typography>
        <Box sx={{ mb: 1 }}>
          <LinearProgress 
            variant="determinate" 
            value={percentage} 
            sx={{ height: 8, borderRadius: 4 }}
          />
        </Box>
        <Typography variant="body2" color="text.secondary">
          {value.toLocaleString()} / {total.toLocaleString()} {unit} ({percentage}%)
        </Typography>
      </CardContent>
    </Card>
  );
};
