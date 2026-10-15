import { Toast } from "../hooks/useToasts";

interface ToastViewportProps {
  toasts: Toast[];
  onDismiss: (id: number) => void;
}

export function ToastViewport({ toasts, onDismiss }: ToastViewportProps) {
  return (
    <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2 w-96 max-w-[calc(100vw-2rem)]">
      {toasts.map((t) => (
        <div
          key={t.id}
          role="alert"
          className={`flex items-start gap-3 rounded-lg shadow-lg border px-4 py-3 text-sm ${
            t.kind === "error" ? "bg-red-50 border-red-200 text-red-800" : "bg-green-50 border-green-200 text-green-800"
          }`}
        >
          <div className="flex-1 min-w-0">
            <p className="font-medium">{t.title}</p>
            <p className="mt-0.5 break-words">{t.message}</p>
          </div>
          <button aria-label="Dismiss" onClick={() => onDismiss(t.id)} className="opacity-60 hover:opacity-100">
            ×
          </button>
        </div>
      ))}
    </div>
  );
}
