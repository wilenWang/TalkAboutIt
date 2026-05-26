import { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { deletePersona } from '../api/client';
import { usePersonaStore } from '../stores/usePersonaStore';
import { useLanguage } from '../i18n/LanguageContext';
import PersonaCard from '../components/PersonaCard';
import ArchetypeFilter from '../components/ArchetypeFilter';

export default function PersonaManagePage() {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const { personas, loading, fetchPersonas } = usePersonaStore();

  const [filterArchetype, setFilterArchetype] = useState<string | null>(null);

  const filteredPersonas = useMemo(() => {
    if (!filterArchetype) return personas;
    return personas.filter((p) => p.archetype === filterArchetype);
  }, [personas, filterArchetype]);

  const handleDelete = async (id: string) => {
    if (!window.confirm(t('msgConfirmDelete'))) return;
    try {
      await deletePersona(id);
      fetchPersonas();
    } catch (e) {
      alert(t('errDeletePersona'));
    }
  };

  return (
    <div className="h-full flex flex-col bg-white">
      <div className="flex-1 overflow-y-auto px-6 py-8 max-w-[900px] mx-auto w-full">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-[22px] font-bold tracking-tight">{t('tabPersonas')}</h2>
          <button
            onClick={() => navigate('/personas/new')}
            className="px-4 py-1.5 bg-[#0075de] text-white text-sm rounded-md hover:bg-[#0066c0] transition-colors"
          >
            {t('actionNewPersona')}
          </button>
        </div>

        <div className="mb-6">
          <ArchetypeFilter active={filterArchetype} onChange={setFilterArchetype} />
        </div>

        {loading ? (
          <div className="text-sm text-[#a39e98]">{t('msgLoading')}</div>
        ) : filteredPersonas.length === 0 ? (
          <div className="text-center py-20">
            <div className="text-4xl mb-3">🔍</div>
            <div className="text-sm text-[#a39e98]">{t('msgNoFilterResults')}</div>
          </div>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
            {filteredPersonas.map((p) => (
              <PersonaCard
                key={p.id}
                persona={p}
                onEdit={(id) => navigate(`/personas/${id}/edit`)}
                onDelete={handleDelete}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
