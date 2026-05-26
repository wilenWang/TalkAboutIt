import { useRef, useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useLanguage, type SystemLanguage } from '../i18n/LanguageContext';
import ConnectionBadge from './ConnectionBadge';
import type { SSEConnectionStatus } from '../hooks/useSSE';
import type { AppStatus } from '../stores/useAppStore';

interface Props {
  title?: string;
  showNav?: boolean;
  sseStatus?: SSEConnectionStatus;
  appStatus?: AppStatus;
}

const systemLanguageOptions: { value: SystemLanguage; label: string; shortLabel: string; flag: string }[] = [
  { value: 'zh-CN', label: '简体中文', shortLabel: '简体中文', flag: '🇨🇳' },
  { value: 'en-US', label: 'English', shortLabel: 'English', flag: '🇺🇸' },
];

export default function Header({ title, showNav = true, sseStatus, appStatus }: Props) {
  const { language: systemLanguage, setLanguage: setSystemLanguage, t } = useLanguage();
  const [languageMenuOpen, setLanguageMenuOpen] = useState(false);
  const languageMenuRef = useRef<HTMLDivElement | null>(null);

  // Close language menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (languageMenuRef.current && !languageMenuRef.current.contains(event.target as Node)) {
        setLanguageMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const currentSystemLanguage = systemLanguageOptions.find((option) => option.value === systemLanguage)!;

  return (
    <header className="px-6 py-3 flex items-center gap-3 border-b border-black/[0.06]">
      <Link to="/" className="text-lg font-bold tracking-tight hover:opacity-80 transition-opacity">
        ✦ TalkAboutIt
      </Link>
      {title && <span className="text-[13px] text-[#a39e98]">{t(title as any)}</span>}
      <span className="flex-1" />

      <div className="relative" ref={languageMenuRef}>
        <button
          type="button"
          onClick={() => setLanguageMenuOpen((open) => !open)}
          className="text-[13px] text-[#615d59] hover:text-black/95 transition-colors"
        >
          🌐 {currentSystemLanguage.shortLabel}
        </button>
        {languageMenuOpen && (
          <div className="absolute right-0 top-full mt-2 min-w-[168px] rounded-xl bg-white shadow-[0_10px_30px_rgba(0,0,0,0.08)] border border-black/[0.06] py-1 z-20">
            {systemLanguageOptions.map((option) => {
              const isActive = option.value === systemLanguage;
              return (
                <button
                  key={option.value}
                  type="button"
                  onClick={() => {
                    setSystemLanguage(option.value);
                    setLanguageMenuOpen(false);
                  }}
                  className={`w-full px-3 py-2 text-left text-[13px] transition-colors ${
                    isActive ? 'bg-[#f2f9ff] text-[#0075de] font-semibold' : 'text-[#615d59] hover:bg-black/[0.03]'
                  }`}
                >
                  {option.flag} {option.label}
                </button>
              );
            })}
          </div>
        )}
      </div>

      {showNav && (
        <Link
          to="/history"
          className="text-[13px] text-[#615d59] hover:text-black/95 transition-colors"
        >
          {t('pageHistory')}
        </Link>
      )}

      {sseStatus !== undefined && appStatus !== undefined && (
        <ConnectionBadge sseStatus={sseStatus} appStatus={appStatus} />
      )}
    </header>
  );
}
