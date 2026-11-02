import { render } from '@testing-library/react';
import { expect, it } from 'vitest';
import { UserIcon } from './UserIcon';

it('renders an svg hidden from assistive tech', () => {
  const { container } = render(<UserIcon size={48} />);
  const svg = container.querySelector('svg.user-icon')!;
  expect(svg).toHaveAttribute('aria-hidden', 'true');
  expect(svg).toHaveAttribute('width', '48');
  expect(svg).toHaveAttribute('height', '48');
});
