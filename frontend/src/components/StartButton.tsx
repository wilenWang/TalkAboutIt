import { useLanguage } from '../i18n/LanguageContext';

interface Props {
  disabled: boolean;
  loading: boolean;
  onClick: () => void;
  hint?: string;
}

export default function StartButton({ disabled, loading, onClick, hint }: Props) {
  const { t } = useLanguage();

  return (
    <div className="flex flex-col items-end gap-1">
      <button
        onClick={onClick}
        disabled={disabled || loading}
        className={`
          px-5 py-2 rounded text-sm font-semibold whitespace-nowrap transition-all
          ${disabled || loading
            ? 'bg-gray-200 text-gray-400 cursor-not-allowed'
            : 'bg-[#0075de] text-white hover:bg-[#0066cc] active:scale-[0.97]'
          }
        `}
      >
        {loading ? t('preparing') : `✦ ${t('startDiscussion')}`}
      </button>
      {hint && (
        <span className="text-[11px] text-[#a39e98]">{hint}</span>
      )}
    </div>
  );
}