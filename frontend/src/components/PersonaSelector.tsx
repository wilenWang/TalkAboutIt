import { useState, useEffect } from 'react';
import { fetchPersonas } from '../api/client';
import type { PersonaSummary } from '../types';
import { useLanguage } from '../i18n/LanguageContext';

type TabKey = 'discussion' | 'personas';

interface Props {
  selected: string[];
  onChange: (selected: string[]) => void;
}

export default function PersonaSelector({ selected, onChange }: Props) {
  const { t, f } = useLanguage();
  const [activeTab, setActiveTab] = useState<TabKey>('discussion');
  const [personas, setPersonas] = useState<PersonaSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');

  useEffect(() => {
    fetchPersonas()
      .then(setPersonas)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const toggle = (id: string) => {
    if (selected.includes(id)) {
      onChange(selected.filter((s) => s !== id));
    } else if (selected.length < 4) {
      onChange([...selected, id]);
    }
  };

  const filteredPersonas = personas.filter((persona) => {
    const keyword = search.trim().toLowerCase();
    if (!keyword) return true;
    const haystack = [
      persona.name,
      persona.display_name,
      persona.role_title,
      persona.description,
      ...persona.tags,
    ]
      .join(' ')
      .toLowerCase();
    return haystack.includes(keyword);
  });

  if (loading) {
    return (
      <div className="w-[260px] bg-[#f6f5f4] border-r border-black/[0.06] flex h-full">
        <div className="w-12 border-r border-black/[0.06] flex flex-col items-center py-3">
          <div className="w-8 h-8 rounded-md bg-black/[0.04]" />
          <div className="w-8 h-8 rounded-md bg-black/[0.04] mt-2" />
        </div>
        <div className="flex-1 p-4 text-sm text-[#a39e98]">{t('msgLoading')}</div>
      </div>
    );
  }

  const tabs: { key: TabKey; icon: string; label: string }[] = [
    { key: 'discussion', icon: '💬', label: t('tabDiscussion') },
    { key: 'personas', icon: '👤', label: t('tabPersonas') },
  ];

  return (
    <div className="w-[260px] bg-[#f6f5f4] border-r border-black/[0.06] flex h-full overflow-hidden">
      {/* Vertical tabs */}
      <div className="w-12 border-r border-black/[0.06] flex flex-col items-center py-2 gap-1">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            title={tab.label}
            className={`w-9 h-9 rounded-md flex items-center justify-center text-lg transition-colors
              ${activeTab === tab.key ? 'bg-white shadow-sm text-black/95' : 'text-[#a39e98] hover:bg-black/[0.04] hover:text-black/70'}
            `}
          >
            {tab.icon}
          </button>
        ))}
      </div>

      {/* Content area */}
      <div className="flex-1 flex flex-col h-full overflow-hidden">
        {/* Discussion tab — persona selection */}
        {activeTab === 'discussion' && (
          <>
            <div className="px-3 pb-3 pt-3">
              <input
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder={t('inputSearch')}
                className="w-full px-2.5 py-2 border border-black/10 rounded text-sm bg-white text-black/95 outline-none focus:border-[#0075de] transition-colors"
              />
            </div>
            <div className="flex-1 overflow-y-auto px-2 pb-2">
              {filteredPersonas.map((p) => {
                const isSelected = selected.includes(p.id);
                return (
                  <div
                    key={p.id}
                    onClick={() => toggle(p.id)}
                    className={`
                      flex items-center gap-2 px-2.5 py-2 rounded-md cursor-pointer transition-colors text-sm font-medium
                      ${isSelected ? 'bg-white shadow-sm border-l-[3px] border-[#0075de] pl-[7px]' : 'hover:bg-black/[0.04]'}
                    `}
                  >
                    <span className="text-xl leading-none flex-shrink-0">{p.avatar}</span>
                    <span className="flex-1 truncate">{p.name}</span>
                    <span
                      className={`
                        w-[18px] h-[18px] rounded flex items-center justify-center text-[10px] transition-all
                        ${isSelected ? 'bg-[#0075de] text-white' : 'border-[1.5px] border-[#a39e98] text-transparent'}
                      `}
                    >
                      ✓
                    </span>
                  </div>
                );
              })}
            </div>
            <div className="p-2 border-t border-black/[0.06]">
              <div className="text-[11px] text-[#a39e98] text-center py-1">
                {selected.length >= 2
                  ? f('fmtSelectedCount', { n: selected.length })
                  : t('msgSelectParticipants')}
              </div>
            </div>
          </>
        )}

        {/* Personas tab — link to management page */}
        {activeTab === 'personas' && (
          <div className="flex-1 flex flex-col items-center justify-center px-6 text-center">
            <div className="text-4xl mb-3">👤</div>
            <div className="text-sm font-medium text-black/70 mb-1">{t('tabPersonas')}</div>
            <div className="text-xs text-[#a39e98] mb-4">
              {t('msgNoPersonas')}
            </div>
            <a
              href="/personas"
              className="px-4 py-2 bg-[#0075de] text-white text-sm rounded-md hover:bg-[#0066c0] transition-colors"
            >
              {t('actionNewPersona')}
            </a>
          </div>
        )}
      </div>
    </div>
  );
}
