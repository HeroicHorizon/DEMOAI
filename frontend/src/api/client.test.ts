import { afterEach, describe, expect, it, vi } from 'vitest';
import { DUPLICATE_EMAIL, SERVER_UNREACHABLE, createUser, getStats, listUsers } from './client';
import { jsonResponse } from '../testUtils';

afterEach(() => vi.unstubAllGlobals());

describe('api client', () => {
  it('maps network failure to the unreachable message', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));
    await expect(getStats()).rejects.toThrow(SERVER_UNREACHABLE);
  });

  it('maps 503 to the unreachable message', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(503, { success: false, error: { message: 'x' } })));
    await expect(getStats()).rejects.toThrow(SERVER_UNREACHABLE);
  });

  it('maps 409 to the duplicate email message', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(409, { success: false, error: { message: 'dup' } })));
    await expect(
      createUser({ firstName: 'a', lastName: 'b', email: 'a@b.co', phone: '', department: '', designation: '' }),
    ).rejects.toThrow(DUPLICATE_EMAIL);
  });

  it('builds list query string', async () => {
    const f = vi.fn().mockResolvedValue(jsonResponse(200, { success: true, data: [], meta: { page: 2, pageSize: 10, total: 0 } }));
    vi.stubGlobal('fetch', f);
    await listUsers({ page: 2, pageSize: 10, search: 'jo', status: 'ACTIVE' });
    const url = String(f.mock.calls[0][0]);
    expect(url).toContain('/api/v1/users?');
    expect(url).toContain('page=2');
    expect(url).toContain('search=jo');
    expect(url).toContain('status=ACTIVE');
  });
});
