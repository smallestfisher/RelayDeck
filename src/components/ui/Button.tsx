import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { cn } from '../../lib/format';

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger' | 'icon';
type ButtonSize = 'sm' | 'md' | 'lg';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  icon?: ReactNode;
}

const variantClass: Record<ButtonVariant, string> = {
  primary:
    'border-primary/80 bg-primary text-white shadow-[0_12px_30px_rgb(37_99_235/0.25)] hover:bg-primary/90',
  secondary:
    'border-line bg-elevated text-text hover:border-primary/45 hover:bg-primary/10',
  ghost: 'border-transparent bg-transparent text-muted hover:bg-elevated hover:text-text',
  danger: 'border-danger/45 bg-danger/10 text-danger hover:bg-danger/15',
  icon: 'border-line bg-elevated text-muted hover:border-primary/45 hover:text-primary',
};

const sizeClass: Record<ButtonSize, string> = {
  sm: 'h-9 px-3 text-sm',
  md: 'h-10 px-4 text-sm',
  lg: 'h-12 px-5 text-base',
};

export function Button({
  variant = 'secondary',
  size = 'md',
  icon,
  className,
  children,
  type = 'button',
  ...props
}: ButtonProps) {
  const isIconOnly = variant === 'icon' && !children;

  return (
    <button
      type={type}
      className={cn(
        'focus-ring inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border font-medium transition disabled:cursor-not-allowed disabled:opacity-55',
        variantClass[variant],
        isIconOnly ? 'h-10 w-10 p-0' : sizeClass[size],
        className
      )}
      {...props}
    >
      {icon}
      {children}
    </button>
  );
}
