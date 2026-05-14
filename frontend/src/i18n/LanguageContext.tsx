import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import { translations, type TranslationKey } from './translations';

export type SystemLanguage = 'zh-CN' | 'en-US';

export interface LanguageContextType {
  language: SystemLanguage;
  setLanguage: (lang: SystemLanguage) => void;
  /** Key-based translation. Accepts a typed key or freeform string (for error messages, etc). */
  t: (key: TranslationKey | string) => string;
  /** Format a pattern translation with dynamic values. Example: f('roundCount', { n: 3 }) */
  f: (key: TranslationKey, params: Record<string, string | number>) => string;
}

const STORAGE_KEY = 'talkaboutit.system-language';

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

function getInitialLanguage(): SystemLanguage {
  if (typeof window === 'undefined') {
    return 'zh-CN';
  }
  const stored = window.localStorage.getItem(STORAGE_KEY);
  return stored === 'en-US' ? 'en-US' : 'zh-CN';
}

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [language, setLanguage] = useState<SystemLanguage>(getInitialLanguage);

  useEffect(() => {
    window.localStorage.setItem(STORAGE_KEY, language);
  }, [language]);

  /** Look up a translation by key. Falls back to key itself if not in table. */
  const t = useCallback(
    (key: TranslationKey | string): string => {
      const row = translations[key as TranslationKey];
      if (!row) return key;
      return language === 'en-US' ? row.en : row.zh;
    },
    [language]
  );

  /** Format a pattern translation. Replaces {param} placeholders. */
  const f = useCallback(
    (key: TranslationKey, params: Record<string, string | number>): string => {
      const row = translations[key];
      const template = row ? (language === 'en-US' ? row.en : row.zh) : key;
      return template.replace(/\{(\w+)\}/g, (_, k) => String(params[k] ?? `{${k}}`));
    },
    [language]
  );

  const value = useMemo(
    () => ({ language, setLanguage, t, f }),
    [language, t, f]
  );

  return <LanguageContext.Provider value={value}>{children}</LanguageContext.Provider>;
}

export function useLanguage() {
  const context = useContext(LanguageContext);
  if (!context) {
    throw new Error('useLanguage must be used within a LanguageProvider');
  }
  return context;
}
