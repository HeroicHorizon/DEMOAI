import { render } from '@testing-library/react';
import type { ReactElement } from 'react';
import { MemoryRouter } from 'react-router-dom';
import { NotificationProvider } from './components/Notification';
import type { User } from './types';

export function renderWithProviders(ui: ReactElement, route = '/') {
  return render(
    <MemoryRouter initialEntries={[route]}>
      <NotificationProvider>{ui}</NotificationProvider>
    </MemoryRouter>,
  );
}

export const makeUser = (n: number, over: Partial<User> = {}): User => ({
  id: `id${n}`,
  firstName: `First${n}`,
  lastName: `Last${n}`,
  email: `user${n}@example.com`,
  phone: '+911234567890',
  department: 'Engineering',
  designation: 'Engineer',
  status: 'ACTIVE',
  createdAt: '2026-10-30T10:00:00Z',
  updatedAt: '2026-10-30T11:00:00Z',
  ...over,
});

export function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
}
