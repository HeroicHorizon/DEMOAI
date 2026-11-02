import { NavLink, Route, Routes } from 'react-router-dom';
import { NotificationProvider } from './components/Notification';
import { Dashboard } from './pages/Dashboard';
import { UserDetails } from './pages/UserDetails';
import { UserForm } from './pages/UserForm';
import { UserList } from './pages/UserList';

export function App() {
  return (
    <NotificationProvider>
      <header className="topbar">
        <strong>User Management</strong>
        <nav>
          <NavLink to="/" end>
            Dashboard
          </NavLink>
          <NavLink to="/users">Users</NavLink>
        </nav>
      </header>
      <main>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/users" element={<UserList />} />
          <Route path="/users/new" element={<UserForm />} />
          <Route path="/users/:id" element={<UserDetails />} />
          <Route path="/users/:id/edit" element={<UserForm />} />
        </Routes>
      </main>
    </NotificationProvider>
  );
}
