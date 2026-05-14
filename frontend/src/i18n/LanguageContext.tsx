import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import { translations, type TranslationKey, type FormatParams } from './translations';

export type SystemLanguage = 'zh-CN' | 'en-US';

export interface LanguageContextType {
  language: SystemLanguage;
  setLanguage: (lang: SystemLanguage) => void;
  /** Key-based translation. Accepts a typed key or freeform string (for error messages, etc). */
  t: (key: TranslationKey | string) => string;
  /** Format a pattern translation with type-checked dynamic values. */
  f: <K extends keyof FormatParams>(key: K, params: FormatParams[K]) => string;
}

const STORAGE_KEY = 'talkaboutit.system-language';

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

function getInitialLanguage(): SystemLanguage {
  if (typeof window === 'undefined') {
    return 'zh-CN';
  }
  const stored = window.localStorage.getItem(STORAGE_KEY);
  return stored === 'en-US' || stored === 'zh-CN' ? stored : 'zh-CN';
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
      return row[language] ?? key;
    },
    [language]
  );

  /** Format a pattern translation. Replaces {param} placeholders. */
  const f = useCallback(
    <K extends keyof FormatParams>(key: K, params: FormatParams[K]): string => {
      const row = translations[key as unknown as TranslationKey];
      const template = row ? (row[language] ?? key) : key;
      return template.replace(/\{(\w+)\}/g, (_, k) => String((params as Record<string, unknown>)[k] ?? `{${k}}`));
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
