import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { getUser } from '../api/client';
import { useNotify } from '../components/Notification';
import type { User } from '../types';

export function UserDetails() {
  const { id } = useParams();
  const [user, setUser] = useState<User | null>(null);
  const notify = useNotify();

  useEffect(() => {
    if (!id) return;
    getUser(id)
      .then(setUser)
      .catch((e: Error) => notify('error', e.message));
  }, [id, notify]);

  if (!user) return <p>Loading…</p>;
  const rows: [string, string][] = [
    ['User ID', user.id],
    ['First Name', user.firstName],
    ['Last Name', user.lastName],
    ['Email Address', user.email],
    ['Phone Number', user.phone],
    ['Department', user.department],
    ['Designation', user.designation],
    ['Status', user.status],
    ['Created Date', new Date(user.createdAt).toLocaleString()],
    ['Updated Date', new Date(user.updatedAt).toLocaleString()],
  ];
  return (
    <section>
      <h1>User Details</h1>
      <dl className="details">
        {rows.map(([k, v]) => (
          <div key={k}>
            <dt>{k}</dt>
            <dd>{v}</dd>
          </div>
        ))}
      </dl>
      <Link to={`/users/${user.id}/edit`}>Edit</Link> <Link to="/users">Back to list</Link>
    </section>
  );
}
