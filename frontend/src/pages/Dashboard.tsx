import { useEffect, useState } from 'react';
import { getStats } from '../api/client';
import { useNotify } from '../components/Notification';
import type { Stats } from '../types';

export function Dashboard() {
  const [stats, setStats] = useState<Stats | null>(null);
  const notify = useNotify();

  useEffect(() => {
    getStats()
      .then(setStats)
      .catch((e: Error) => notify('error', e.message));
  }, [notify]);

  return (
    <section>
      <h1>Dashboard</h1>
      <div className="cards">
        <div className="card">
          <h2>Total users</h2>
          <p data-testid="total-count">{stats ? stats.total : '–'}</p>
        </div>
        <div className="card">
          <h2>Active users</h2>
          <p data-testid="active-count">{stats ? stats.active : '–'}</p>
        </div>
        <div className="card">
          <h2>Inactive users</h2>
          <p data-testid="inactive-count">{stats ? stats.inactive : '–'}</p>
        </div>
      </div>
    </section>
  );
}
