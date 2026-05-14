import { useLanguage, type SystemLanguage } from '../i18n/LanguageContext';

interface Props {
  value: SystemLanguage;
  onChange: (value: SystemLanguage) => void;
}

export default function LanguageToggle({ value, onChange }: Props) {
  const { t } = useLanguage();
  const options: { value: SystemLanguage; label: string; flag: string }[] = [
    { value: 'zh-CN', label: '中文', flag: '🇨🇳' },
    { value: 'en-US', label: 'English', flag: '🇺🇸' },
  ];

  return (
    <div className="flex flex-col gap-1">
      <label className="text-[11px] font-semibold text-[#a39e98] uppercase tracking-wider">
        {t('语言', 'Language')}
      </label>
      <div className="flex gap-1">
        {options.map((opt) => (
          <button
            key={opt.value}
            type="button"
            onClick={() => onChange(opt.value)}
            className={`px-3 py-2 rounded text-[13px] font-medium transition-colors ${
              value === opt.value
                ? 'bg-[#0075de] text-white shadow-sm'
                : 'bg-gray-100 text-gray-500 hover:bg-gray-200'
            }`}
          >
            {opt.flag} {opt.label}
          </button>
        ))}
      </div>
    </div>
  );
}
