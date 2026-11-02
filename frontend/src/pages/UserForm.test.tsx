import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Route, Routes } from 'react-router-dom';
import { afterEach, expect, it, vi } from 'vitest';
import { jsonResponse, makeUser, renderWithProviders } from '../testUtils';
import { UserForm } from './UserForm';

afterEach(() => vi.unstubAllGlobals());

const renderForm = (route = '/users/new') =>
  renderWithProviders(
    <Routes>
      <Route path="/users/new" element={<UserForm />} />
      <Route path="/users/:id/edit" element={<UserForm />} />
      <Route path="/users" element={<p>list page</p>} />
      <Route path="/users/:id" element={<p>details page</p>} />
    </Routes>,
    route,
  );

async function fillValid() {
  const user = userEvent.setup();
  await user.type(screen.getByLabelText('First Name'), 'John');
  await user.type(screen.getByLabelText('Last Name'), 'Doe');
  await user.type(screen.getByLabelText('Email'), 'john@example.com');
  return user;
}

it('shows required-field errors and sends no request', async () => {
  const f = vi.fn();
  vi.stubGlobal('fetch', f);
  renderForm();
  await userEvent.click(screen.getByRole('button', { name: 'Save' }));
  expect(screen.getByText('First name is required.')).toBeInTheDocument();
  expect(screen.getByText('Last name is required.')).toBeInTheDocument();
  expect(screen.getByText('Email is required.')).toBeInTheDocument();
  expect(f).not.toHaveBeenCalled();
});

it('shows an invalid email error and sends no request', async () => {
  const f = vi.fn();
  vi.stubGlobal('fetch', f);
  renderForm();
  const user = userEvent.setup();
  await user.type(screen.getByLabelText('First Name'), 'John');
  await user.type(screen.getByLabelText('Last Name'), 'Doe');
  await user.type(screen.getByLabelText('Email'), 'not-an-email');
  await user.click(screen.getByRole('button', { name: 'Save' }));
  expect(screen.getByText('Email address is invalid.')).toBeInTheDocument();
  expect(f).not.toHaveBeenCalled();
});

it('submits a valid form to the API and navigates to the list', async () => {
  const f = vi.fn().mockResolvedValue(jsonResponse(201, { success: true, data: makeUser(1) }));
  vi.stubGlobal('fetch', f);
  renderForm();
  const user = await fillValid();
  await user.click(screen.getByRole('button', { name: 'Save' }));
  await waitFor(() => expect(screen.getByText('list page')).toBeInTheDocument());
  const [url, init] = f.mock.calls[0];
  expect(String(url)).toContain('/api/v1/users');
  expect(init.method).toBe('POST');
  expect(JSON.parse(init.body)).toMatchObject({ firstName: 'John', email: 'john@example.com' });
});

it('shows the duplicate email message on a 409', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(409, { success: false, error: { message: 'dup' } })));
  renderForm();
  const user = await fillValid();
  await user.click(screen.getByRole('button', { name: 'Save' }));
  const msgs = await screen.findAllByText('A user with this email already exists.');
  expect(msgs.length).toBeGreaterThan(0);
  expect(screen.queryByText('list page')).not.toBeInTheDocument();
});

it('loads an existing user and updates it with PUT', async () => {
  const f = vi
    .fn()
    .mockResolvedValueOnce(jsonResponse(200, { success: true, data: makeUser(3) }))
    .mockResolvedValueOnce(jsonResponse(200, { success: true, data: makeUser(3) }));
  vi.stubGlobal('fetch', f);
  renderForm('/users/id3/edit');
  await waitFor(() => expect(screen.getByLabelText('First Name')).toHaveValue('First3'));
  await userEvent.click(screen.getByRole('button', { name: 'Save' }));
  await waitFor(() => expect(screen.getByText('details page')).toBeInTheDocument());
  expect(f.mock.calls[1][1].method).toBe('PUT');
  expect(String(f.mock.calls[1][0])).toContain('/api/v1/users/id3');
});
