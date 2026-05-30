import { useEffect, useState } from 'react';
import type { Quote } from '../../types';
import { Button } from '../ui/Button';
import './QuoteCard.css';

interface QuoteCardProps {
  quote: Quote;
  onDismiss?: () => void;
}

export function QuoteCard({ quote, onDismiss }: QuoteCardProps) {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setVisible(true), 50);
    return () => clearTimeout(timer);
  }, []);

  return (
    <div className={`quote-card ${visible ? 'quote-card-visible' : ''}`}>
      <div className="quote-card-mark">&ldquo;</div>
      <blockquote className="quote-card-text">{quote.text}</blockquote>
      <cite className="quote-card-author">— {quote.author}</cite>
      {onDismiss && (
        <Button variant="ghost" size="sm" onClick={onDismiss} className="quote-card-dismiss">
          Dismiss
        </Button>
      )}
    </div>
  );
}
