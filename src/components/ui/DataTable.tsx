import type { ReactNode } from 'react';
import { cn } from '../../lib/format';

interface DataTableProps {
  children: ReactNode;
  className?: string;
}

export function DataTable({ children, className }: DataTableProps) {
  return (
    <div className={cn('thin-scrollbar overflow-x-auto rounded-lg border border-line bg-panel/45', className)}>
      <table className="table-grid min-w-full text-left text-sm">{children}</table>
    </div>
  );
}

export const tableHeadClass = 'bg-elevated/80 text-xs font-medium text-muted';
export const tableCellClass = 'whitespace-nowrap px-4 py-3 align-middle';
