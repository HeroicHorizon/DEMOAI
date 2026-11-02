import type { ListParams } from '../types';

interface Props {
  value: ListParams['status'];
  onChange: (v: ListParams['status']) => void;
}

export function StatusFilter({ value, onChange }: Props) {
  return (
    <select aria-label="Status filter" value={value} onChange={(e) => onChange(e.target.value as Props['value'])}>
      <option value="">All statuses</option>
      <option value="ACTIVE">ACTIVE</option>
      <option value="INACTIVE">INACTIVE</option>
    </select>
  );
}
