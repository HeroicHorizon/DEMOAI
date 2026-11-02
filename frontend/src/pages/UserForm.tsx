import { useEffect, useState, type FormEvent } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { ApiError, createUser, getUser, updateUser } from '../api/client';
import { useNotify } from '../components/Notification';
import type { UserInput } from '../types';
import { validateUser, type FormErrors } from '../validation';

const EMPTY: UserInput = { firstName: '', lastName: '', email: '', phone: '', department: '', designation: '' };

const FIELDS: { key: keyof UserInput; label: string; type?: string; required?: boolean }[] = [
  { key: 'firstName', label: 'First Name', required: true },
  { key: 'lastName', label: 'Last Name', required: true },
  { key: 'email', label: 'Email', type: 'text', required: true },
  { key: 'phone', label: 'Phone' },
  { key: 'department', label: 'Department' },
  { key: 'designation', label: 'Designation' },
];

export function UserForm() {
  const { id } = useParams();
  const navigate = useNavigate();
  const notify = useNotify();
  const [values, setValues] = useState<UserInput>(EMPTY);
  const [errors, setErrors] = useState<FormErrors>({});
  const [formError, setFormError] = useState('');
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!id) return;
    getUser(id)
      .then(({ firstName, lastName, email, phone, department, designation }) =>
        setValues({ firstName, lastName, email, phone, department, designation }),
      )
      .catch((e: Error) => notify('error', e.message));
  }, [id, notify]);

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setFormError('');
    const found = validateUser(values);
    setErrors(found);
    if (Object.keys(found).length > 0) return;
    setBusy(true);
    try {
      if (id) await updateUser(id, values);
      else await createUser(values);
      notify('success', id ? 'User updated.' : 'User created.');
      navigate(id ? `/users/${id}` : '/users');
    } catch (err) {
      const apiErr = err as ApiError;
      if (apiErr.status === 409) setErrors({ email: apiErr.message });
      else if (apiErr.status === 400 && Object.keys(apiErr.details ?? {}).length > 0) setErrors(apiErr.details);
      else setFormError(apiErr.message);
      notify('error', apiErr.message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <section>
      <h1>{id ? 'Edit User' : 'Add User'}</h1>
      <form onSubmit={submit} noValidate className="form">
        {FIELDS.map((f) => (
          <div key={f.key} className="field">
            <label htmlFor={f.key}>{f.label}</label>
            <input
              id={f.key}
              type={f.type ?? 'text'}
              value={values[f.key]}
              aria-invalid={!!errors[f.key]}
              onChange={(e) => setValues((v) => ({ ...v, [f.key]: e.target.value }))}
            />
            {errors[f.key] && <span className="field-error">{errors[f.key]}</span>}
          </div>
        ))}
        {formError && <p className="field-error">{formError}</p>}
        <div className="actions">
          <button type="button" onClick={() => navigate(-1)}>
            Cancel
          </button>
          <button type="submit" disabled={busy}>
            Save
          </button>
        </div>
      </form>
    </section>
  );
}
