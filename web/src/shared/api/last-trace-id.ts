type Listener = (traceId: string | null) => void;

let current: string | null = null;
const listeners = new Set<Listener>();

export function recordTraceId(traceId: string | null): void {
  current = traceId;
  for (const fn of listeners) fn(current);
}

export function getLastTraceId(): string | null {
  return current;
}

export function subscribeLastTraceId(fn: Listener): () => void {
  listeners.add(fn);
  return () => {
    listeners.delete(fn);
  };
}
