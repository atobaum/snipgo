// Import Wails generated bindings
import * as WailsApp from '../wailsjs/go/app/App';
import { core } from '../wailsjs/go/models';
import { Snippet } from './types';

// Convert Wails core.Snippet to our Snippet type
function convertSnippet(wailsSnippet: core.Snippet): Snippet {
  return {
    id: wailsSnippet.id,
    title: wailsSnippet.title,
    tags: wailsSnippet.tags,
    language: wailsSnippet.language,
    is_favorite: wailsSnippet.is_favorite,
    created_at: typeof wailsSnippet.created_at === 'string' 
      ? wailsSnippet.created_at 
      : new Date(wailsSnippet.created_at).toISOString(),
    updated_at: typeof wailsSnippet.updated_at === 'string'
      ? wailsSnippet.updated_at
      : new Date(wailsSnippet.updated_at).toISOString(),
    body: wailsSnippet.body,
  };
}

// Convert our Snippet to Wails core.Snippet
function convertToWailsSnippet(snippet: Snippet): core.Snippet {
  return core.Snippet.createFrom({
    id: snippet.id,
    title: snippet.title,
    tags: snippet.tags,
    language: snippet.language,
    is_favorite: snippet.is_favorite,
    created_at: snippet.created_at,
    updated_at: snippet.updated_at,
    body: snippet.body,
  });
}

// Re-export with our Snippet type
export interface App {
  GetAllSnippets(): Promise<Snippet[]>;
  GetSnippet(id: string): Promise<Snippet>;
  SaveSnippet(snippet: Snippet): Promise<void>;
  DeleteSnippet(id: string): Promise<void>;
  SearchSnippets(query: string): Promise<Snippet[]>;
  CopyToClipboard(text: string): Promise<void>;
  ReloadSnippets(): Promise<void>;
}

// Use Wails generated bindings with type conversion
export const app: App = {
  GetAllSnippets: async () => {
    const result = await WailsApp.GetAllSnippets();
    return result.map(convertSnippet);
  },
  GetSnippet: async (id: string) => {
    const result = await WailsApp.GetSnippet(id);
    return convertSnippet(result);
  },
  SaveSnippet: async (snippet: Snippet) => {
    await WailsApp.SaveSnippet(convertToWailsSnippet(snippet));
  },
  DeleteSnippet: WailsApp.DeleteSnippet,
  SearchSnippets: async (query: string) => {
    const result = await WailsApp.SearchSnippets(query);
    return result.map(convertSnippet);
  },
  CopyToClipboard: WailsApp.CopyToClipboard,
  ReloadSnippets: WailsApp.ReloadSnippets,
};
