import { useCallback, useEffect, useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { deleteUser, listUsers, setUserStatus } from '../api/client';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { useNotify } from '../components/Notification';
import { Pagination } from '../components/Pagination';
import { StatusFilter } from '../components/StatusFilter';
import { UsersTable } from '../components/UsersTable';
import type { ListParams, ListResult, User } from '../types';

export const PAGE_SIZE = 10;

export function UserList() {
  const [params, setParams] = useState<ListParams>({ page: 1, pageSize: PAGE_SIZE, search: '', status: '' });
  const [searchText, setSearchText] = useState('');
  const [result, setResult] = useState<ListResult>({ users: [], meta: { page: 1, pageSize: PAGE_SIZE, total: 0 } });
  const [toDelete, setToDelete] = useState<User | null>(null);
  const notify = useNotify();

  const load = useCallback(() => {
    listUsers(params)
      .then(setResult)
      .catch((e: Error) => notify('error', e.message));
  }, [params, notify]);

  useEffect(load, [load]);

  const onSearch = (e: FormEvent) => {
    e.preventDefault();
    setParams((p) => ({ ...p, search: searchText.trim(), page: 1 }));
  };

  const toggle = async (u: User) => {
    const next = u.status === 'ACTIVE' ? 'INACTIVE' : 'ACTIVE';
    try {
      await setUserStatus(u.id, next);
      notify('success', `User ${next === 'ACTIVE' ? 'activated' : 'deactivated'}.`);
      load();
    } catch (e) {
      notify('error', (e as Error).message);
    }
  };

  const confirmDelete = async () => {
    if (!toDelete) return;
    const u = toDelete;
    setToDelete(null);
    try {
      await deleteUser(u.id);
      notify('success', 'User deleted.');
      load();
    } catch (e) {
      notify('error', (e as Error).message);
    }
  };

  return (
    <section>
      <div className="toolbar">
        <h1>User Management</h1>
        <Link to="/users/new" className="button">
          Add User
        </Link>
      </div>
      <form className="filters" onSubmit={onSearch} role="search">
        <input
          type="search"
          aria-label="Search users"
          placeholder="Search by name or email"
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
        />
        <button type="submit">Search</button>
        <StatusFilter value={params.status} onChange={(status) => setParams((p) => ({ ...p, status, page: 1 }))} />
      </form>
      <UsersTable users={result.users} onToggleStatus={toggle} onDelete={setToDelete} />
      <Pagination
        page={result.meta.page}
        pageSize={result.meta.pageSize}
        total={result.meta.total}
        onPage={(page) => setParams((p) => ({ ...p, page }))}
      />
      {toDelete && (
        <ConfirmDialog
          message={`Delete ${toDelete.firstName} ${toDelete.lastName}? This cannot be undone.`}
          onConfirm={confirmDelete}
          onCancel={() => setToDelete(null)}
        />
      )}
    </section>
  );
}
