import type { PersonaSummary } from '../types';
import { useLanguage } from '../i18n/LanguageContext';
import { isImageAvatar } from '../utils/avatar';

interface Props {
  persona: PersonaSummary;
  onEdit: (id: string) => void;
  onDelete: (id: string) => void;
}

export default function PersonaCard({ persona, onEdit, onDelete }: Props) {
  const { t, language } = useLanguage();
  const label = language === 'zh-CN' ? persona.display_name : persona.name;

  return (
    <div className="group relative bg-white rounded-xl aspect-square flex flex-col items-center justify-center border border-black/[0.06] cursor-pointer overflow-hidden hover:shadow-[rgba(0,0,0,0.04)_0px_4px_18px] transition-shadow">
      {/* Avatar */}
      {isImageAvatar(persona.avatar) ? (
        <img
          src={persona.avatar}
          alt={label}
          className="w-16 h-16 rounded-xl object-cover mb-2"
        />
      ) : (
        <span className="text-[40px] mb-1">{persona.avatar}</span>
      )}

      {/* Name */}
      <div className="font-semibold text-sm text-black/95">{label}</div>

      {/* Archetype */}
      {persona.archetype && (
        <div className="text-[11px] text-[#a39e98] mt-1">
          {t('archetype' + persona.archetype) || persona.archetype}
        </div>
      )}

      {/* Hover overlay */}
      <div className="absolute inset-0 bg-white/[0.98] flex flex-col items-center justify-center p-3.5 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
        <div className="text-xs text-[#615d59] text-center line-clamp-5 leading-relaxed">
          {persona.description}
        </div>
        <div className="mt-2.5 flex gap-1.5">
          <button
            onClick={(e) => {
              e.stopPropagation();
              onEdit(persona.id);
            }}
            className="px-2.5 py-1 bg-[#f2f9ff] text-[#0075de] rounded text-[11px] hover:bg-[#e0f0ff] transition-colors"
          >
            {t('actionEdit')}
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onDelete(persona.id);
            }}
            className="px-2.5 py-1 bg-red-50 text-red-500 rounded text-[11px] hover:bg-red-100 transition-colors"
          >
            {t('actionDelete')}
          </button>
        </div>
      </div>
    </div>
  );
}
