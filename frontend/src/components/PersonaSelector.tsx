import { useState, useEffect } from 'react';
import { fetchPersonas } from '../api/client';
import type { PersonaSummary } from '../types';
import { useLanguage } from '../i18n/LanguageContext';

interface Props {
  selected: string[];
  onChange: (selected: string[]) => void;
}

export default function PersonaSelector({ selected, onChange }: Props) {
  const { t, f } = useLanguage();
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
      <div className="w-[260px] bg-[#f6f5f4] border-r border-black/[0.06] flex flex-col h-full">
        <div className="p-4 text-sm text-[#a39e98]">{t('loading')}</div>
      </div>
    );
  }

  return (
    <div className="w-[260px] bg-[#f6f5f4] border-r border-black/[0.06] flex flex-col h-full overflow-hidden">
      <div className="px-4 pt-4 pb-2 text-[11px] font-semibold text-[#a39e98] uppercase tracking-wider">
        {t('participants')}
      </div>
      <div className="px-4 pb-3">
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder={t('search')}
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
            ? f('selectedCount', { n: selected.length })
            : t('selectParticipants')}
        </div>
      </div>
    </div>
  );
}