import { ClipboardList } from 'lucide-react';
import { Button } from '../components/ui/Button';

interface EmptyPageProps {
  title: string;
  description: string;
  onReturn: () => void;
}

export function EmptyPage({ title, description, onReturn }: EmptyPageProps) {
  return (
    <section className="flex min-h-[calc(100vh-140px)] items-center justify-center">
      <div className="surface-card max-w-xl rounded-xl p-10 text-center">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/12 text-primary">
          <ClipboardList className="h-8 w-8" />
        </div>
        <h1 className="mt-6 text-2xl font-semibold text-text">{title}</h1>
        <p className="mt-3 text-sm leading-7 text-muted">{description}</p>
        <Button className="mt-6" onClick={onReturn}>
          返回概览
        </Button>
      </div>
    </section>
  );
}
