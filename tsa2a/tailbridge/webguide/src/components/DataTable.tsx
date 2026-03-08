import React from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Typography,
  Box,
  TablePagination,
} from '@mui/material';

interface Column<T> {
  key: keyof T | string;
  label: string;
  render?: (item: T) => React.ReactNode;
  align?: 'left' | 'right' | 'center';
}

interface DataTableProps<T> {
  data: T[];
  columns: Column<T>[];
  title?: string;
  emptyMessage?: string;
  pagination?: boolean;
  page?: number;
  rowsPerPage?: number;
  onPageChange?: (page: number) => void;
  onRowsPerPageChange?: (rowsPerPage: number) => void;
  totalRows?: number;
}

export function DataTable<T extends Record<string, unknown>>({
  data,
  columns,
  title,
  emptyMessage = 'No data available',
  pagination = false,
  page = 0,
  rowsPerPage = 10,
  onPageChange,
  onRowsPerPageChange,
  totalRows,
}: DataTableProps<T>) {
  const getValue = (item: T, key: string): unknown => {
    return key.split('.').reduce((obj, prop) => {
      return obj && typeof obj === 'object' ? (obj as Record<string, unknown>)[prop] : undefined;
    }, item as unknown);
  };

  return (
    <Paper sx={{ width: '100%', overflow: 'hidden' }}>
      {title && (
        <Box sx={{ p: 2, borderBottom: 1, borderColor: 'divider' }}>
          <Typography variant="h6">{title}</Typography>
        </Box>
      )}
      <TableContainer sx={{ maxHeight: pagination ? 440 : 'none' }}>
        <Table stickyHeader>
          <TableHead>
            <TableRow>
              {columns.map((column) => (
                <TableCell
                  key={String(column.key)}
                  align={column.align || 'left'}
                  sx={{ minWidth: column.key === 'actions' ? 100 : undefined }}
                >
                  {column.label}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {data.length === 0 ? (
              <TableRow>
                <TableCell colSpan={columns.length} align="center" sx={{ py: 4 }}>
                  <Typography color="text.secondary">{emptyMessage}</Typography>
                </TableCell>
              </TableRow>
            ) : (
              data.map((item, idx) => (
                <TableRow key={idx} hover>
                  {columns.map((column) => (
                    <TableCell key={String(column.key)} align={column.align || 'left'}>
                      {column.render
                        ? column.render(item)
                        : String(getValue(item, String(column.key)) ?? '')}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>
      {pagination && onPageChange && (
        <TablePagination
          component="div"
          count={totalRows ?? data.length}
          page={page}
          onPageChange={(_, newPage) => onPageChange(newPage)}
          rowsPerPage={rowsPerPage}
          onRowsPerPageChange={
            onRowsPerPageChange
              ? (event) => onRowsPerPageChange(parseInt(event.target.value, 10))
              : undefined
          }
          rowsPerPageOptions={onRowsPerPageChange ? [5, 10, 25, 50] : [10]}
        />
      )}
    </Paper>
  );
}

interface StatusChipProps {
  status: string;
  mapping?: Record<string, 'success' | 'error' | 'warning' | 'info'>;
}

export const StatusChip: React.FC<StatusChipProps> = ({ status, mapping }) => {
  const defaultMapping: Record<string, 'success' | 'error' | 'warning' | 'info'> = {
    online: 'success',
    active: 'success',
    completed: 'success',
    sending: 'info',
    pending: 'warning',
    offline: 'error',
    failed: 'error',
    cancelled: 'error',
    idle: 'warning',
  };

  const colorMap = mapping || defaultMapping;
  const color = colorMap[status.toLowerCase()] || 'info';

  return (
    <Chip
      label={status}
      color={color}
      size="small"
      sx={{ textTransform: 'capitalize' }}
    />
  );
};
