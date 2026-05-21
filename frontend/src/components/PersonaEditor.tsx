import { useState, useEffect } from 'react';
import { fetchPersona, createPersona, updatePersona } from '../api/client';
import type { Persona } from '../types/persona';
import { createEmptyPersona } from '../types/persona';
import { useLanguage } from '../i18n/LanguageContext';

interface Props {
  personaId: string | null;
  onSave: () => void;
  onCancel: () => void;
}

function StringArrayInput({
  label,
  values,
  onChange,
}: {
  label: string;
  values: string[];
  onChange: (v: string[]) => void;
}) {
  const [text, setText] = useState(values.join('\n'));
  useEffect(() => {
    setText(values.join('\n'));
  }, [values]);
  return (
    <div>
      <label className="block text-xs font-medium text-[#615d59] mb-1">{label}</label>
      <textarea
        value={text}
        onChange={(e) => {
          setText(e.target.value);
          onChange(e.target.value.split('\n').filter((s) => s.trim()));
        }}
        rows={3}
        className="w-full px-2.5 py-2 border border-black/10 rounded text-sm bg-white text-black/95 outline-none focus:border-[#0075de] transition-colors resize-none"
        placeholder="每行一个"
      />
    </div>
  );
}

function TextInput({
  label,
  value,
  onChange,
  placeholder,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
}) {
  return (
    <div>
      <label className="block text-xs font-medium text-[#615d59] mb-1">{label}</label>
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full px-2.5 py-2 border border-black/10 rounded text-sm bg-white text-black/95 outline-none focus:border-[#0075de] transition-colors"
      />
    </div>
  );
}

function NumberInput({
  label,
  value,
  onChange,
  min,
  max,
}: {
  label: string;
  value: number;
  onChange: (v: number) => void;
  min: number;
  max: number;
}) {
  return (
    <div>
      <label className="block text-xs font-medium text-[#615d59] mb-1">{label}</label>
      <input
        type="number"
        min={min}
        max={max}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        className="w-full px-2.5 py-2 border border-black/10 rounded text-sm bg-white text-black/95 outline-none focus:border-[#0075de] transition-colors"
      />
    </div>
  );
}

function TextArea({
  label,
  value,
  onChange,
  placeholder,
  rows = 3,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  rows?: number;
}) {
  return (
    <div>
      <label className="block text-xs font-medium text-[#615d59] mb-1">{label}</label>
      <textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        rows={rows}
        className="w-full px-2.5 py-2 border border-black/10 rounded text-sm bg-white text-black/95 outline-none focus:border-[#0075de] transition-colors resize-none"
      />
    </div>
  );
}

export default function PersonaEditor({ personaId, onSave, onCancel }: Props) {
  const { t } = useLanguage();
  const [p, setP] = useState<Persona>(createEmptyPersona());
  const [loading, setLoading] = useState(personaId !== null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (personaId) {
      fetchPersona(personaId)
        .then(setP)
        .catch(console.error)
        .finally(() => setLoading(false));
    }
  }, [personaId]);

  const update = <K extends keyof Persona>(key: K, value: Persona[K]) => {
    setP((prev) => ({ ...prev, [key]: value }));
  };

  const handleSave = async () => {
    if (!p.id.trim() || !p.name.trim()) {
      alert('ID 和名称不能为空');
      return;
    }
    setSaving(true);
    try {
      if (personaId) {
        await updatePersona(personaId, p);
      } else {
        await createPersona(p);
      }
      onSave();
    } catch (e) {
      alert(personaId ? t('errUpdatePersona') : t('errCreatePersona'));
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="h-screen flex items-center justify-center bg-white">
        <div className="text-sm text-[#a39e98]">{t('msgLoading')}</div>
      </div>
    );
  }

  return (
    <div className="h-screen flex flex-col bg-white">
      {/* Header */}
      <header className="px-6 py-3 flex items-center gap-3 border-b border-black/[0.06]">
        <button
          onClick={onCancel}
          className="text-[13px] text-[#615d59] hover:text-black/95 transition-colors"
        >
          ← {t('actionCancel')}
        </button>
        <span className="text-lg font-bold tracking-tight">✦ TalkAboutIt</span>
        <span className="flex-1" />
        <button
          onClick={handleSave}
          disabled={saving}
          className="px-5 py-1.5 bg-[#0075de] text-white text-sm rounded-md hover:bg-[#0066c0] transition-colors disabled:opacity-50"
        >
          {saving ? '...' : t('actionSave')}
        </button>
      </header>

      {/* Form */}
      <main className="flex-1 overflow-y-auto px-6 py-8 max-w-[720px] mx-auto w-full">
        <h2 className="text-[22px] font-bold tracking-tight mb-6">
          {personaId ? t('actionEdit') : t('actionNewPersona')}
        </h2>

        <div className="space-y-6">
          {/* Basic Info */}
          <section>
            <h3 className="text-sm font-semibold text-black/95 mb-3">{t('labelName')}</h3>
            <div className="grid grid-cols-2 gap-3">
              <TextInput
                label="ID"
                value={p.id}
                onChange={(v) => update('id', v)}
                placeholder="unique-id"
              />
              <TextInput
                label={t('labelName')}
                value={p.name}
                onChange={(v) => update('name', v)}
              />
              <TextInput
                label={t('labelDisplayName')}
                value={p.display_name}
                onChange={(v) => update('display_name', v)}
              />
              <TextInput
                label={t('labelAvatar')}
                value={p.avatar}
                onChange={(v) => update('avatar', v)}
                placeholder="emoji"
              />
              <TextInput
                label={t('labelRoleTitle')}
                value={p.role_title}
                onChange={(v) => update('role_title', v)}
              />
              <TextInput
                label="Archetype"
                value={p.archetype}
                onChange={(v) => update('archetype', v)}
                placeholder="e.g. Visionary, Engineer, Philosopher"
              />
            </div>
            <div className="mt-3">
              <TextArea
                label={t('labelDescription')}
                value={p.description}
                onChange={(v) => update('description', v)}
                rows={3}
              />
            </div>
            <div className="mt-3">
              <StringArrayInput
                label={t('labelTags')}
                values={p.tags}
                onChange={(v) => update('tags', v)}
              />
            </div>
          </section>

          {/* Stance */}
          <section>
            <h3 className="text-sm font-semibold text-black/95 mb-3">立场</h3>
            <div className="space-y-3">
              <TextArea
                label="默认立场"
                value={p.stance.default_position}
                onChange={(v) => update('stance', { ...p.stance, default_position: v })}
              />
              <NumberInput
                label="强度 (1-5)"
                value={p.stance.intensity}
                onChange={(v) => update('stance', { ...p.stance, intensity: v })}
                min={1}
                max={5}
              />
              <StringArrayInput
                label="偏见"
                values={p.stance.biases}
                onChange={(v) => update('stance', { ...p.stance, biases: v })}
              />
              <StringArrayInput
                label="禁忌"
                values={p.stance.taboos}
                onChange={(v) => update('stance', { ...p.stance, taboos: v })}
              />
            </div>
          </section>

          {/* Core Beliefs */}
          <section>
            <h3 className="text-sm font-semibold text-black/95 mb-3">核心信念</h3>
            <div className="space-y-3">
              {p.core_beliefs.map((cb, i) => (
                <div key={i} className="border border-black/10 rounded-lg p-3 space-y-2">
                  <TextInput
                    label="信念"
                    value={cb.belief}
                    onChange={(v) => {
                      const copy = [...p.core_beliefs];
                      copy[i] = { ...cb, belief: v };
                      update('core_beliefs', copy);
                    }}
                  />
                  <NumberInput
                    label="优先级 (1-5)"
                    value={cb.priority}
                    onChange={(v) => {
                      const copy = [...p.core_beliefs];
                      copy[i] = { ...cb, priority: v };
                      update('core_beliefs', copy);
                    }}
                    min={1}
                    max={5}
                  />
                  <TextArea
                    label="理由"
                    value={cb.rationale}
                    onChange={(v) => {
                      const copy = [...p.core_beliefs];
                      copy[i] = { ...cb, rationale: v };
                      update('core_beliefs', copy);
                    }}
                    rows={2}
                  />
                  <button
                    onClick={() => update('core_beliefs', p.core_beliefs.filter((_, idx) => idx !== i))}
                    className="text-xs text-red-500 hover:underline"
                  >
                    删除
                  </button>
                </div>
              ))}
              <button
                onClick={() =>
                  update('core_beliefs', [...p.core_beliefs, { belief: '', priority: 3, rationale: '' }])
                }
                className="text-sm text-[#0075de] hover:underline"
              >
                + 添加信念
              </button>
            </div>
          </section>

          {/* Speaking Style */}
          <section>
            <h3 className="text-sm font-semibold text-black/95 mb-3">说话风格</h3>
            <div className="grid grid-cols-2 gap-3">
              <TextInput
                label="语气"
                value={p.speaking_style.tone}
                onChange={(v) => update('speaking_style', { ...p.speaking_style, tone: v })}
              />
              <TextInput
                label="节奏"
                value={p.speaking_style.cadence}
                onChange={(v) => update('speaking_style', { ...p.speaking_style, cadence: v })}
              />
              <NumberInput
                label=" verbosity (1-5)"
                value={p.speaking_style.verbosity}
                onChange={(v) => update('speaking_style', { ...p.speaking_style, verbosity: v })}
                min={1}
                max={5}
              />
            </div>
            <div className="mt-3 space-y-3">
              <StringArrayInput
                label="标志性表达"
                values={p.speaking_style.signature_patterns}
                onChange={(v) => update('speaking_style', { ...p.speaking_style, signature_patterns: v })}
              />
              <StringArrayInput
                label="Do"
                values={p.speaking_style.do}
                onChange={(v) => update('speaking_style', { ...p.speaking_style, do: v })}
              />
              <StringArrayInput
                label="Don't"
                values={p.speaking_style.dont}
                onChange={(v) => update('speaking_style', { ...p.speaking_style, dont: v })}
              />
            </div>
          </section>

          {/* Knowledge Scope */}
          <section>
            <h3 className="text-sm font-semibold text-black/95 mb-3">知识范围</h3>
            <div className="space-y-3">
              <StringArrayInput
                label="领域"
                values={p.knowledge_scope.domains}
                onChange={(v) => update('knowledge_scope', { ...p.knowledge_scope, domains: v })}
              />
              <TextInput
                label="时间截止"
                value={p.knowledge_scope.time_cutoff}
                onChange={(v) => update('knowledge_scope', { ...p.knowledge_scope, time_cutoff: v })}
              />
              <TextArea
                label="未知处理方式"
                value={p.knowledge_scope.unknown_handling}
                onChange={(v) => update('knowledge_scope', { ...p.knowledge_scope, unknown_handling: v })}
                rows={2}
              />
              <StringArrayInput
                label="禁止声明"
                values={p.knowledge_scope.forbidden_claims}
                onChange={(v) => update('knowledge_scope', { ...p.knowledge_scope, forbidden_claims: v })}
              />
            </div>
          </section>

          {/* Interaction Rules */}
          <section>
            <h3 className="text-sm font-semibold text-black/95 mb-3">互动规则</h3>
            <div className="grid grid-cols-2 gap-3">
              <TextInput
                label="称呼他人"
                value={p.interaction_rules.address_others}
                onChange={(v) => update('interaction_rules', { ...p.interaction_rules, address_others: v })}
              />
              <TextInput
                label="分歧风格"
                value={p.interaction_rules.disagreement_style}
                onChange={(v) => update('interaction_rules', { ...p.interaction_rules, disagreement_style: v })}
              />
              <TextInput
                label="打断策略"
                value={p.interaction_rules.interruption_policy}
                onChange={(v) => update('interaction_rules', { ...p.interaction_rules, interruption_policy: v })}
              />
              <TextInput
                label="提问策略"
                value={p.interaction_rules.question_policy}
                onChange={(v) => update('interaction_rules', { ...p.interaction_rules, question_policy: v })}
              />
              <TextInput
                label="让步策略"
                value={p.interaction_rules.concession_policy}
                onChange={(v) => update('interaction_rules', { ...p.interaction_rules, concession_policy: v })}
              />
            </div>
            <div className="mt-3">
              <StringArrayInput
                label="避免"
                values={p.interaction_rules.avoid}
                onChange={(v) => update('interaction_rules', { ...p.interaction_rules, avoid: v })}
              />
            </div>
          </section>

          {/* Debate Goal */}
          <section>
            <h3 className="text-sm font-semibold text-black/95 mb-3">辩论目标</h3>
            <div className="space-y-3">
              <TextArea
                label="主要目标"
                value={p.debate_goal.primary_goal}
                onChange={(v) => update('debate_goal', { ...p.debate_goal, primary_goal: v })}
                rows={2}
              />
              <StringArrayInput
                label="次要目标"
                values={p.debate_goal.secondary_goals}
                onChange={(v) => update('debate_goal', { ...p.debate_goal, secondary_goals: v })}
              />
              <TextArea
                label="胜利条件"
                value={p.debate_goal.win_condition}
                onChange={(v) => update('debate_goal', { ...p.debate_goal, win_condition: v })}
                rows={2}
              />
              <TextArea
                label="失败条件"
                value={p.debate_goal.loss_condition}
                onChange={(v) => update('debate_goal', { ...p.debate_goal, loss_condition: v })}
                rows={2}
              />
            </div>
          </section>

          {/* Prompting */}
          <section>
            <h3 className="text-sm font-semibold text-black/95 mb-3">提示词</h3>
            <div className="space-y-3">
              <TextArea
                label="系统前言"
                value={p.prompting.system_preamble}
                onChange={(v) => update('prompting', { ...p.prompting, system_preamble: v })}
                rows={3}
              />
              <StringArrayInput
                label="回复约束"
                values={p.prompting.reply_constraints}
                onChange={(v) => update('prompting', { ...p.prompting, reply_constraints: v })}
              />
            </div>
          </section>

          {/* Examples */}
          <section>
            <h3 className="text-sm font-semibold text-black/95 mb-3">示例</h3>
            <div className="space-y-3">
              <TextArea
                label="开场白"
                value={p.examples.opening_line}
                onChange={(v) => update('examples', { ...p.examples, opening_line: v })}
                rows={2}
              />
              <TextArea
                label="反驳示例"
                value={p.examples.sample_rebuttal}
                onChange={(v) => update('examples', { ...p.examples, sample_rebuttal: v })}
                rows={2}
              />
            </div>
          </section>
        </div>

        <div className="h-12" />
      </main>
    </div>
  );
}
