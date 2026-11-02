export type UserStatus = 'ACTIVE' | 'INACTIVE';

export interface User {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  department: string;
  designation: string;
  status: UserStatus;
  createdAt: string;
  updatedAt: string;
}

export interface UserInput {
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  department: string;
  designation: string;
}

export interface Stats {
  total: number;
  active: number;
  inactive: number;
}

export interface ListParams {
  page: number;
  pageSize: number;
  search: string;
  status: '' | UserStatus;
}

export interface ListResult {
  users: User[];
  meta: { page: number; pageSize: number; total: number };
}
