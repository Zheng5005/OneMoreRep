import { api } from './client';
import type { Quote } from '../types';

export const quoteApi = {
  random: (): Promise<Quote> =>
    api.get<Quote>('/quotes/random'),
};
