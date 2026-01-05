import { useEffect, useRef, useCallback } from "react";

/**
 * Hook for debounced save operations with cancel and flush capabilities
 */
export function useDebouncedSave<T extends (...args: unknown[]) => void>(
  callback: T,
  delay: number
): {
  debouncedSave: (...args: Parameters<T>) => void;
  cancelPendingSave: () => void;
  flushSave: () => void;
} {
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const callbackRef = useRef(callback);
  const pendingArgsRef = useRef<Parameters<T> | null>(null);

  // Keep callback ref up to date
  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const cancelPendingSave = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    pendingArgsRef.current = null;
  }, []);

  const flushSave = useCallback(() => {
    if (timeoutRef.current && pendingArgsRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
      callbackRef.current(...pendingArgsRef.current);
      pendingArgsRef.current = null;
    }
  }, []);

  const debouncedSave = useCallback(
    (...args: Parameters<T>) => {
      pendingArgsRef.current = args;

      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }

      timeoutRef.current = setTimeout(() => {
        callbackRef.current(...args);
        pendingArgsRef.current = null;
        timeoutRef.current = null;
      }, delay);
    },
    [delay]
  );

  return { debouncedSave, cancelPendingSave, flushSave };
}
