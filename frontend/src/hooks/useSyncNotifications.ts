import { useEffect, useRef } from "react";
import { onSyncStatus } from "../bridge";
import { Toast } from "./useToasts";

export function useSyncNotifications(push: (toast: Omit<Toast, "id">) => void) {
  const lastAutoError = useRef<string | null>(null);
  const pushRef = useRef(push);
  pushRef.current = push;

  useEffect(() => {
    return onSyncStatus((status) => {
      if (status.ok) {
        lastAutoError.current = null;
        if (status.manual) {
          pushRef.current({ kind: "success", title: "Synced", message: status.summary });
        }
        return;
      }
      const key = `${status.errorKind}:${status.message}`;
      if (!status.manual && lastAutoError.current === key) {
        return;
      }
      if (!status.manual) {
        lastAutoError.current = key;
      }
      pushRef.current({ kind: "error", title: "Sync failed", message: status.message });
    });
  }, []);
}
