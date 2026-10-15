export interface Snippet {
  id: string;
  title: string;
  description: string;
  tags: string[];
  language: string;
  is_favorite: boolean;
  created_at: string;
  updated_at: string;
  body: string;
}

export interface Variable {
  name: string;
  description?: string;
  default?: string;
  choices?: string[];
}

export interface SyncStatus {
  lastRunAt: string;
  ok: boolean;
  manual: boolean;
  summary: string;
  errorKind: string;
  message: string;
  conflictFiles: string[];
  paused: boolean;
}

export type SyncConflict = "manual" | "local" | "remote";

export interface SyncSettings {
  enabled: boolean;
  intervalMinutes: number;
  conflict: SyncConflict;
}


