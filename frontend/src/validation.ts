import type { UserInput } from './types';

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const PHONE_RE = /^\+?[0-9][0-9 -]{5,19}$/;

export type FormErrors = Partial<Record<keyof UserInput, string>>;

export function validateUser(v: UserInput): FormErrors {
  const e: FormErrors = {};
  if (!v.firstName.trim()) e.firstName = 'First name is required.';
  if (!v.lastName.trim()) e.lastName = 'Last name is required.';
  if (!v.email.trim()) e.email = 'Email is required.';
  else if (!EMAIL_RE.test(v.email.trim())) e.email = 'Email address is invalid.';
  if (v.phone.trim() && !PHONE_RE.test(v.phone.trim())) e.phone = 'Phone number is invalid.';
  if (v.firstName.length > 100) e.firstName = 'First name is too long.';
  if (v.lastName.length > 100) e.lastName = 'Last name is too long.';
  if (v.department.length > 100) e.department = 'Department is too long.';
  if (v.designation.length > 100) e.designation = 'Designation is too long.';
  return e;
}
