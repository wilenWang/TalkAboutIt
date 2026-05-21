import { useLanguage } from '../i18n/LanguageContext';

const ARCHETYPES = ['Visionary', 'Engineer', 'Philosopher', 'Craftsman', 'Operator'] as const;

interface Props {
  active: string | null;
  onChange: (archetype: string | null) => void;
}

export default function ArchetypeFilter({ active, onChange }: Props) {
  const { t } = useLanguage();

  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-[13px] text-[#615d59]">{t('labelFilter')}:</span>
      <button
        onClick={() => onChange(null)}
        className={`px-3 py-1 rounded-full text-[13px] transition-colors ${
          active === null
            ? 'border border-[#0075de] bg-[#f2f9ff] text-[#0075de]'
            : 'border border-black/10 bg-white text-[#615d59] hover:bg-black/[0.03]'
        }`}
      >
        {t('labelFilterAll')}
      </button>
      {ARCHETYPES.map((arch) => (
        <button
          key={arch}
          onClick={() => onChange(arch)}
          className={`px-3 py-1 rounded-full text-[13px] transition-colors ${
            active === arch
              ? 'border border-[#0075de] bg-[#f2f9ff] text-[#0075de]'
              : 'border border-black/10 bg-white text-[#615d59] hover:bg-black/[0.03]'
          }`}
        >
          {t('archetype' + arch) || arch}
        </button>
      ))}
    </div>
  );
}
