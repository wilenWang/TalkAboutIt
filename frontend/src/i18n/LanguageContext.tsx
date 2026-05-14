import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';

export type SystemLanguage = 'zh-CN' | 'en-US';

export interface LanguageContextType {
  language: SystemLanguage;
  setLanguage: (lang: SystemLanguage) => void;
  t: (zh: string, en: string) => string;
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

  const t = useCallback(
    (zh: string, en: string) => (language === 'en-US' ? en : zh),
    [language]
  );

  const value = useMemo(
    () => ({
      language,
      setLanguage,
      t,
    }),
    [language, t]
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
