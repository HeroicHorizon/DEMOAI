import { createContext, useCallback, useContext, useState, type ReactNode } from 'react';

type Kind = 'success' | 'error';
interface Note {
  id: number;
  kind: Kind;
  text: string;
}
interface Ctx {
  notify: (kind: Kind, text: string) => void;
}

const NotifyContext = createContext<Ctx>({ notify: () => {} });
export const useNotify = () => useContext(NotifyContext).notify;

let nextId = 1;

export function NotificationProvider({ children }: { children: ReactNode }) {
  const [notes, setNotes] = useState<Note[]>([]);
  const notify = useCallback((kind: Kind, text: string) => {
    const id = nextId++;
    setNotes((n) => [...n, { id, kind, text }]);
    setTimeout(() => setNotes((n) => n.filter((x) => x.id !== id)), 5000);
  }, []);
  return (
    <NotifyContext.Provider value={{ notify }}>
      {children}
      <div className="notifications" aria-live="polite">
        {notes.map((n) => (
          <div key={n.id} role={n.kind === 'error' ? 'alert' : 'status'} className={`note ${n.kind}`}>
            {n.text}
          </div>
        ))}
      </div>
    </NotifyContext.Provider>
  );
}
