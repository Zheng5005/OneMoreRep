import './Spinner.css';

interface SpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  color?: string;
  className?: string;
}

export function Spinner({ size = 'md', color, className = '' }: SpinnerProps) {
  return (
    <div
      className={`spinner spinner-${size} ${className}`}
      style={color ? { color } : undefined}
      role="status"
      aria-label="Loading"
    >
      <div className="spinner-circle" />
    </div>
  );
}
