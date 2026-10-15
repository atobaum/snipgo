import { useCallback, useEffect, useRef, useState } from "react";

export interface Toast {
  id: number;
  kind: "error" | "success";
  title: string;
  message: string;
}

const SUCCESS_TOAST_MS = 4000;

export function useToasts() {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const nextId = useRef(1);
  const timers = useRef(new Map<number, ReturnType<typeof setTimeout>>());

  const dismiss = useCallback((id: number) => {
    const timer = timers.current.get(id);
    if (timer) {
      clearTimeout(timer);
      timers.current.delete(id);
    }
    setToasts((ts) => ts.filter((t) => t.id !== id));
  }, []);

  const push = useCallback(
    (toast: Omit<Toast, "id">) => {
      const id = nextId.current++;
      setToasts((ts) => [...ts, { ...toast, id }]);
      if (toast.kind === "success") {
        timers.current.set(id, setTimeout(() => dismiss(id), SUCCESS_TOAST_MS));
      }
    },
    [dismiss]
  );

  useEffect(() => {
    const pending = timers.current;
    return () => pending.forEach(clearTimeout);
  }, []);

  return { toasts, push, dismiss };
}
