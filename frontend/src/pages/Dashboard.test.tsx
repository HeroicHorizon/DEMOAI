import { screen, waitFor } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { jsonResponse, renderWithProviders } from '../testUtils';
import { Dashboard } from './Dashboard';

afterEach(() => vi.unstubAllGlobals());

it('shows total, active and inactive counts', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(200, { success: true, data: { total: 7, active: 5, inactive: 2 } })));
  renderWithProviders(<Dashboard />);
  await waitFor(() => expect(screen.getByTestId('total-count')).toHaveTextContent('7'));
  expect(screen.getByTestId('active-count')).toHaveTextContent('5');
  expect(screen.getByTestId('inactive-count')).toHaveTextContent('2');
});

it('shows the unreachable notification when the server is down', async () => {
  vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('fail')));
  renderWithProviders(<Dashboard />);
  expect(await screen.findByText('Unable to reach the server. Please try again later.')).toBeInTheDocument();
});
