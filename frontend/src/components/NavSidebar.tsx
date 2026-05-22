import { useLanguage } from '../i18n/LanguageContext';

type NavPage = 'talk' | 'personas';

interface Props {
  active: NavPage;
  onNavigate: (page: NavPage) => void;
}

const NAV_ITEMS: { key: NavPage; icon: string }[] = [
  { key: 'talk', icon: '💬' },
  { key: 'personas', icon: '👤' },
];

export default function NavSidebar({ active, onNavigate }: Props) {
  const { t } = useLanguage();

  const labelMap: Record<NavPage, string> = {
    talk: t('tabDiscussion'),
    personas: t('tabPersonas'),
  };

  return (
    <div className="w-[72px] bg-[#f6f5f4] border-r border-black/[0.06] flex flex-col items-center py-4 gap-1 flex-shrink-0">
      <div className="text-lg font-bold tracking-tight mb-4">✦</div>
      {NAV_ITEMS.map((item) => (
        <button
          key={item.key}
          onClick={() => onNavigate(item.key)}
          className={`w-[56px] py-2 rounded-lg flex flex-col items-center gap-0.5 transition-colors ${
            active === item.key
              ? 'bg-white shadow-sm text-[#0075de]'
              : 'text-[#a39e98] hover:bg-black/[0.04] hover:text-black/70'
          }`}
        >
          <span className="text-lg">{item.icon}</span>
          <span className="text-[10px] leading-tight">{labelMap[item.key]}</span>
        </button>
      ))}
    </div>
  );
}
