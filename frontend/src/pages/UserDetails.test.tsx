import { screen } from '@testing-library/react';
import { Route, Routes } from 'react-router-dom';
import { afterEach, expect, it, vi } from 'vitest';
import { jsonResponse, makeUser, renderWithProviders } from '../testUtils';
import { UserDetails } from './UserDetails';

afterEach(() => vi.unstubAllGlobals());

it('shows all user fields', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(200, { success: true, data: makeUser(4) })));
  renderWithProviders(
    <Routes>
      <Route path="/users/:id" element={<UserDetails />} />
    </Routes>,
    '/users/id4',
  );
  expect(await screen.findByText('user4@example.com')).toBeInTheDocument();
  for (const v of ['id4', 'First4', 'Last4', '+911234567890', 'Engineering', 'Engineer', 'ACTIVE']) {
    expect(screen.getByText(v)).toBeInTheDocument();
  }
  for (const k of ['Created Date', 'Updated Date']) expect(screen.getByText(k)).toBeInTheDocument();
});

it('shows an error notification for a missing user', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(404, { success: false, error: { message: 'User not found.' } })));
  renderWithProviders(
    <Routes>
      <Route path="/users/:id" element={<UserDetails />} />
    </Routes>,
    '/users/5f8d0d55b54764421b7156c3',
  );
  expect(await screen.findByText('User not found.')).toBeInTheDocument();
});

it('shows the user icon', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(200, { success: true, data: makeUser(4) })));
  const { container } = renderWithProviders(
    <Routes>
      <Route path="/users/:id" element={<UserDetails />} />
    </Routes>,
    '/users/id4',
  );
  expect(await screen.findByText('First4 Last4')).toBeInTheDocument();
  expect(container.querySelectorAll('svg.user-icon')).toHaveLength(1);
});
