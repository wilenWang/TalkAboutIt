import { useState, useEffect, useCallback, useMemo } from 'react';
import { fetchPersonas, deletePersona } from '../api/client';
import type { PersonaSummary } from '../types';
import { useLanguage } from '../i18n/LanguageContext';
import PersonaEditor from '../components/PersonaEditor';
import PersonaCard from '../components/PersonaCard';
import ArchetypeFilter from '../components/ArchetypeFilter';

interface Props {
  onBack: () => void;
}

export default function PersonaManagePage({ onBack }: Props) {
  const { t } = useLanguage();
  const [personas, setPersonas] = useState<PersonaSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [filterArchetype, setFilterArchetype] = useState<string | null>(null);

  const load = useCallback(() => {
    setLoading(true);
    fetchPersonas()
      .then(setPersonas)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const filteredPersonas = useMemo(() => {
    if (!filterArchetype) return personas;
    return personas.filter((p) => p.archetype === filterArchetype);
  }, [personas, filterArchetype]);

  const handleDelete = async (id: string) => {
    if (!window.confirm(t('msgConfirmDelete'))) return;
    try {
      await deletePersona(id);
      load();
    } catch (e) {
      alert(t('errDeletePersona'));
    }
  };

  if (editingId || isCreating) {
    return (
      <PersonaEditor
        personaId={editingId}
        onSave={() => {
          setEditingId(null);
          setIsCreating(false);
          load();
        }}
        onCancel={() => {
          setEditingId(null);
          setIsCreating(false);
        }}
      />
    );
  }

  return (
    <div className="h-screen flex flex-col bg-white">
      {/* Header */}
      <header className="px-6 py-3 flex items-center gap-3 border-b border-black/[0.06]">
        <button
          onClick={onBack}
          className="text-[13px] text-[#615d59] hover:text-black/95 transition-colors"
        >
          ← {t('actionBack')}
        </button>
        <span className="text-lg font-bold tracking-tight">✦ TalkAboutIt</span>
        <span className="flex-1" />
        <button
          onClick={() => setIsCreating(true)}
          className="px-4 py-1.5 bg-[#0075de] text-white text-sm rounded-md hover:bg-[#0066c0] transition-colors"
        >
          {t('actionNewPersona')}
        </button>
      </header>

      {/* Content */}
      <main className="flex-1 overflow-y-auto px-6 py-8 max-w-[900px] mx-auto w-full">
        <h2 className="text-[22px] font-bold tracking-tight mb-6">{t('tabPersonas')}</h2>

        {/* Filter */}
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
                onEdit={setEditingId}
                onDelete={handleDelete}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
