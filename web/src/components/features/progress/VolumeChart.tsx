import { useEffect, useState } from 'react';
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
} from 'recharts';
import { useStore } from '../../../stores';
import { Button } from '../../ui/Button';
import './VolumeChart.css';

type GroupBy = 'session' | 'week' | 'month';

interface VolumeChartProps {
  exerciseId?: string;
}

interface ChartDataPoint {
  period: string;
  total_volume: number;
}

export function VolumeChart({ exerciseId }: VolumeChartProps) {
  const { volumeData, loading, error, fetchVolume } = useStore();
  const [groupBy, setGroupBy] = useState<GroupBy>('week');

  useEffect(() => {
    fetchVolume(groupBy, exerciseId);
  }, [groupBy, exerciseId, fetchVolume]);

  const formatXAxis = (periodValue: string | number) => {
    const period = String(periodValue);
    const date = new Date(period);
    if (groupBy === 'session') {
      return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    }
    if (groupBy === 'week') {
      return `Week of ${date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}`;
    }
    return date.toLocaleDateString('en-US', { month: 'short', year: '2-digit' });
  };

  if (loading && volumeData.length === 0) {
    return (
      <div className="volume-chart-loading">
        <div className="loading-spinner" />
        <p>Loading volume data...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="volume-chart-error">
        <p>{error}</p>
        <Button variant="secondary" onClick={() => fetchVolume(groupBy, exerciseId)}>
          Retry
        </Button>
      </div>
    );
  }

  const chartData: ChartDataPoint[] = volumeData.map(v => ({
    period: v.period || v.session_id || '',
    total_volume: v.total_volume,
  }));

  return (
    <div className="volume-chart">
      <div className="volume-chart-header">
        <h3>Volume Over Time</h3>
        <div className="volume-chart-controls">
          <Button
            variant={groupBy === 'session' ? 'primary' : 'ghost'}
            size="sm"
            onClick={() => setGroupBy('session')}
          >
            Session
          </Button>
          <Button
            variant={groupBy === 'week' ? 'primary' : 'ghost'}
            size="sm"
            onClick={() => setGroupBy('week')}
          >
            Week
          </Button>
          <Button
            variant={groupBy === 'month' ? 'primary' : 'ghost'}
            size="sm"
            onClick={() => setGroupBy('month')}
          >
            Month
          </Button>
        </div>
      </div>

      {chartData.length === 0 ? (
        <div className="volume-chart-empty">
          <div className="empty-icon">📈</div>
          <h4>No volume data yet</h4>
          <p>Complete some workouts to see your volume trends.</p>
        </div>
      ) : (
        <div className="volume-chart-container">
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
              <defs>
                <linearGradient id="volumeGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="var(--color-primary)" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="var(--color-primary)" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
              <XAxis
                dataKey="period"
                tickFormatter={formatXAxis}
                stroke="var(--color-text-muted)"
                fontSize={12}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                tickFormatter={(v) => `${(v / 1000).toFixed(0)}k`}
                stroke="var(--color-text-muted)"
                fontSize={12}
                tickLine={false}
                axisLine={false}
                width={50}
              />
              <Tooltip
                formatter={(value) => [`${Number(value).toLocaleString()} lbs`, 'Volume']}
                contentStyle={{
                  background: 'var(--color-surface)',
                  border: '1px solid var(--color-border)',
                  borderRadius: 'var(--radius-md)',
                }}
              />
              <Area
                type="monotone"
                dataKey="total_volume"
                stroke="var(--color-primary)"
                strokeWidth={2}
                fill="url(#volumeGradient)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}