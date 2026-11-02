import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, it, vi } from 'vitest';
import { jsonResponse, makeUser, renderWithProviders } from '../testUtils';
import { UserList } from './UserList';

afterEach(() => vi.unstubAllGlobals());

const listBody = (users: unknown[], total = users.length, page = 1) =>
  jsonResponse(200, { success: true, data: users, meta: { page, pageSize: 10, total } });

it('renders columns and rows', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(listBody([makeUser(1), makeUser(2, { status: 'INACTIVE' })])));
  renderWithProviders(<UserList />);
  expect(await screen.findAllByTestId('user-row')).toHaveLength(2);
  for (const h of ['User ID', 'First Name', 'Last Name', 'Email Address', 'Phone Number', 'Department', 'Designation', 'Status', 'Created Date', 'Updated Date']) {
    expect(screen.getByRole('columnheader', { name: h })).toBeInTheDocument();
  }
  expect(screen.getByText('user1@example.com')).toBeInTheDocument();
  expect(screen.getByRole('button', { name: 'Activate' })).toBeInTheDocument();
  expect(screen.getByRole('button', { name: 'Deactivate' })).toBeInTheDocument();
});

it('shows the empty state with zero data rows', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(listBody([])));
  renderWithProviders(<UserList />);
  expect(await screen.findByText('No users found.')).toBeInTheDocument();
  expect(screen.queryAllByTestId('user-row')).toHaveLength(0);
});

it('sends search and status filter as query parameters', async () => {
  const f = vi.fn().mockResolvedValue(listBody([]));
  vi.stubGlobal('fetch', f);
  renderWithProviders(<UserList />);
  const user = userEvent.setup();
  await user.type(screen.getByLabelText('Search users'), 'john');
  await user.click(screen.getByRole('button', { name: 'Search' }));
  await waitFor(() => expect(String(f.mock.calls.at(-1)![0])).toContain('search=john'));
  await user.selectOptions(screen.getByLabelText('Status filter'), 'ACTIVE');
  await waitFor(() => {
    const url = String(f.mock.calls.at(-1)![0]);
    expect(url).toContain('status=ACTIVE');
    expect(url).toContain('search=john');
  });
  await user.selectOptions(screen.getByLabelText('Status filter'), 'INACTIVE');
  await waitFor(() => expect(String(f.mock.calls.at(-1)![0])).toContain('status=INACTIVE'));
});

it('requests the next page from the server', async () => {
  const f = vi.fn().mockImplementation((url: string) =>
    Promise.resolve(String(url).includes('page=2') ? listBody([makeUser(11)], 11, 2) : listBody([makeUser(1)], 11, 1)),
  );
  vi.stubGlobal('fetch', f);
  renderWithProviders(<UserList />);
  await screen.findByText('user1@example.com');
  await userEvent.click(screen.getByRole('button', { name: 'Next' }));
  expect(await screen.findByText('user11@example.com')).toBeInTheDocument();
  expect(String(f.mock.calls.at(-1)![0])).toContain('page=2');
});

it('deactivates an active user with PATCH', async () => {
  const f = vi.fn().mockImplementation((url: string, init?: RequestInit) =>
    Promise.resolve(init?.method === 'PATCH' ? jsonResponse(200, { success: true, data: makeUser(1, { status: 'INACTIVE' }) }) : listBody([makeUser(1)])),
  );
  vi.stubGlobal('fetch', f);
  renderWithProviders(<UserList />);
  await userEvent.click(await screen.findByRole('button', { name: 'Deactivate' }));
  await waitFor(() => {
    const patch = f.mock.calls.find((c) => c[1]?.method === 'PATCH')!;
    expect(String(patch[0])).toContain('/api/v1/users/id1/status');
    expect(JSON.parse(patch[1].body)).toEqual({ status: 'INACTIVE' });
  });
});

it('sends no DELETE when the confirmation is cancelled', async () => {
  const f = vi.fn().mockResolvedValue(listBody([makeUser(1)]));
  vi.stubGlobal('fetch', f);
  renderWithProviders(<UserList />);
  await userEvent.click(await screen.findByRole('button', { name: 'Delete' }));
  const dialog = screen.getByRole('dialog');
  await userEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }));
  expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  expect(f.mock.calls.some((c) => c[1]?.method === 'DELETE')).toBe(false);
});

it('sends DELETE after confirmation', async () => {
  const f = vi.fn().mockImplementation((url: string, init?: RequestInit) =>
    Promise.resolve(init?.method === 'DELETE' ? jsonResponse(200, { success: true, data: { deleted: true } }) : listBody([makeUser(1)])),
  );
  vi.stubGlobal('fetch', f);
  renderWithProviders(<UserList />);
  await userEvent.click(await screen.findByRole('button', { name: 'Delete' }));
  await userEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Confirm' }));
  await waitFor(() => {
    const del = f.mock.calls.find((c) => c[1]?.method === 'DELETE')!;
    expect(String(del[0])).toContain('/api/v1/users/id1');
  });
});

it('shows the unreachable notification when the list fails', async () => {
  vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('fail')));
  renderWithProviders(<UserList />);
  expect(await screen.findByText('Unable to reach the server. Please try again later.')).toBeInTheDocument();
});

it('renders a user icon in each row', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(listBody([makeUser(1), makeUser(2)])));
  const { container } = renderWithProviders(<UserList />);
  const rows = await screen.findAllByTestId('user-row');
  expect(container.querySelectorAll('svg.user-icon')).toHaveLength(rows.length);
  expect(screen.getByRole('columnheader', { name: 'First Name' })).toBeInTheDocument();
});
