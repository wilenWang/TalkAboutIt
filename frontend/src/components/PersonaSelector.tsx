import { useState, useEffect, useMemo } from 'react';
import { fetchPersonas } from '../api/client';
import type { PersonaSummary } from '../types';
import { useLanguage } from '../i18n/LanguageContext';
import ArchetypeFilter from './ArchetypeFilter';

interface Props {
  selected: string[];
  onChange: (selected: string[]) => void;
}

function isImageAvatar(avatar: string): boolean {
  return avatar.startsWith('http') || avatar.includes('/');
}

export default function PersonaSelector({ selected, onChange }: Props) {
  const { t, f } = useLanguage();
  const [personas, setPersonas] = useState<PersonaSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [filterArchetype, setFilterArchetype] = useState<string | null>(null);

  useEffect(() => {
    fetchPersonas()
      .then(setPersonas)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const filteredPersonas = useMemo(() => {
    if (!filterArchetype) return personas;
    return personas.filter((p) => p.archetype === filterArchetype);
  }, [personas, filterArchetype]);

  const toggle = (id: string) => {
    if (selected.includes(id)) {
      onChange(selected.filter((s) => s !== id));
    } else if (selected.length < 4) {
      onChange([...selected, id]);
    }
  };

  if (loading) {
    return (
      <div className="w-[20%] flex-shrink-0 bg-[#f6f5f4] border-r border-black/[0.06] flex flex-col h-full">
        <div className="flex-1 flex items-center justify-center text-sm text-[#a39e98]">
          {t('msgLoading')}
        </div>
      </div>
    );
  }

  return (
    <div className="w-[20%] flex-shrink-0 bg-[#f6f5f4] border-r border-black/[0.06] flex flex-col h-full overflow-hidden">
      {/* Header */}
      <div className="px-3 py-2.5 border-b border-black/[0.06]">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-bold text-black/95">{t('labelParticipants')}</h3>
          <span className="text-xs text-[#0075de] font-medium">
            {f('fmtSelectedCount', { n: selected.length })}
          </span>
        </div>
        <ArchetypeFilter active={filterArchetype} onChange={setFilterArchetype} />
      </div>

      {/* Grid */}
      <div className="flex-1 overflow-y-auto p-2">
        {filteredPersonas.length === 0 ? (
          <div className="text-center py-10">
            <div className="text-2xl mb-2">🔍</div>
            <div className="text-xs text-[#a39e98]">{t('msgNoFilterResults')}</div>
          </div>
        ) : (
          <div className="grid grid-cols-3 gap-1.5">
            {filteredPersonas.map((p) => {
              const isSelected = selected.includes(p.id);
              return (
                <div
                  key={p.id}
                  onClick={() => toggle(p.id)}
                  className={`
                    relative rounded-lg flex flex-col items-center justify-center cursor-pointer transition-all
                    aspect-square
                    ${isSelected
                      ? 'bg-white border-2 border-[#0075de] shadow-sm'
                      : 'bg-white border border-black/[0.06] hover:shadow-[rgba(0,0,0,0.04)_0px_4px_18px]'
                    }
                  `}
                >
                  {/* Avatar */}
                  {isImageAvatar(p.avatar) ? (
                    <img src={p.avatar} alt={p.name} className="w-9 h-9 rounded-lg object-cover" />
                  ) : (
                    <span className="text-[26px] leading-none">{p.avatar}</span>
                  )}

                  {/* Name */}
                  <div className="text-[10px] font-semibold text-black/95 truncate w-full text-center mt-1">
                    {p.name}
                  </div>

                  {/* Selected check */}
                  {isSelected && (
                    <div className="absolute top-0.5 right-0.5 w-4 h-4 bg-[#0075de] rounded-full flex items-center justify-center">
                      <span className="text-white text-[8px]">✓</span>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
