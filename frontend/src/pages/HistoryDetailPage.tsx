import { useEffect, useState, useMemo } from 'react';
import { getRoundtable } from '../api/client';
import type { PersonaSummary, ReplayMessage } from '../types';
import MessageCard from '../components/MessageCard';
import { useLanguage } from '../i18n/LanguageContext';

interface Props {
  id: string;
  personaList: PersonaSummary[];
  onBack: () => void;
}

export default function HistoryDetailPage({ id, personaList, onBack }: Props) {
  const { language, t } = useLanguage();
  const [topic, setTopic] = useState('');
  const [messages, setMessages] = useState<ReplayMessage[]>([]);
  const [participants, setParticipants] = useState<string[]>([]);
  const [status, setStatus] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    getRoundtable(id)
      .then((snap) => {
        setTopic(snap.topic);
        setParticipants(snap.personas);
        setStatus(snap.status);
        const mapped: ReplayMessage[] = snap.messages.map((m) => {
          const matched = personaList.find((p) => p.id === m.persona_id);
          return {
            id: m.id,
            avatar: matched?.avatar ?? '🤖',
            author: matched?.name ?? m.persona_id,
            personaId: m.persona_id,
            round: m.round,
            content: m.content,
          };
        });
        setMessages(mapped);
        setLoading(false);
      })
      .catch((e) => {
        setError(e instanceof Error ? e.message : t('加载回放失败', 'Failed to load replay'));
        setLoading(false);
      });
  }, [id, personaList, t]);

  // 按轮次分组
  const grouped = useMemo(() => {
    const map = new Map<number, ReplayMessage[]>();
    messages.forEach((m) => {
      const arr = map.get(m.round) ?? [];
      arr.push(m);
      map.set(m.round, arr);
    });
    const rounds = Array.from(map.keys()).sort((a, b) => a - b);
    return rounds.map((round) => ({
      round,
      messages: map.get(round)!,
    }));
  }, [messages]);

  const getPersonaNames = (ids: string[]) =>
    ids
      .map((personaId) => personaList.find((p) => p.id === personaId)?.name ?? personaId)
      .join(language === 'en-US' ? ', ' : '、');

  const formatStatus = (value: string) => {
    switch (value) {
      case 'pending':
        return t('待开始', 'Pending');
      case 'running':
        return t('进行中', 'Running');
      case 'completed':
        return t('已完成', 'Completed');
      case 'failed':
        return t('失败', 'Failed');
      default:
        return value;
    }
  };

  const formatParticipants = () => getPersonaNames(participants);

  return (
    <div className="min-h-screen bg-white">
      {/* Header */}
      <header className="px-6 py-3 flex items-center gap-3 border-b border-black/[0.06]">
        <button
          onClick={onBack}
          className="text-sm text-[#615d59] hover:text-black/95 transition-colors"
        >
          ← {t('返回列表', 'Back to list')}
        </button>
        <span className="text-lg font-bold tracking-tight">✦ TalkAboutIt</span>
        <span className="text-[13px] text-[#a39e98]">{t('回放', 'Replay')}</span>
      </header>

      <main className="max-w-[720px] mx-auto px-6 py-8">
        {/* 加载态 */}
        {loading && (
          <div className="space-y-4 animate-pulse">
            <div className="h-5 bg-black/[0.06] rounded w-1/2" />
            <div className="h-4 bg-black/[0.04] rounded w-full" />
            <div className="h-4 bg-black/[0.04] rounded w-5/6" />
            <div className="h-4 bg-black/[0.04] rounded w-full" />
          </div>
        )}

        {/* 错误态 */}
        {!loading && error && (
          <div className="text-center py-12">
            <div className="text-3xl mb-2">⚠️</div>
            <p className="text-sm text-red-500 mb-3">{error}</p>
            <button
              onClick={() => {
                setLoading(true);
                setError(null);
                getRoundtable(id)
                  .then((snap) => {
                    setTopic(snap.topic);
                    setParticipants(snap.personas);
                    setStatus(snap.status);
                    const mapped: ReplayMessage[] = snap.messages.map((m) => {
                      const matched = personaList.find((p) => p.id === m.persona_id);
                      return {
                        id: m.id,
                        avatar: matched?.avatar ?? '🤖',
                        author: matched?.name ?? m.persona_id,
                        personaId: m.persona_id,
                        round: m.round,
                        content: m.content,
                      };
                    });
                    setMessages(mapped);
                  })
                  .catch((e) => setError(e instanceof Error ? e.message : t('加载回放失败', 'Failed to load replay')))
                  .finally(() => setLoading(false));
              }}
              className="px-4 py-1.5 rounded text-sm font-semibold bg-[#0075de] text-white hover:bg-[#0066cc]"
            >
              {t('重试', 'Retry')}
            </button>
          </div>
        )}

        {/* 回放内容 */}
        {!loading && !error && (
          <>
            <div className="mb-6">
              <h2 className="text-[22px] font-bold tracking-tight mb-1">{topic || t('讨论回放', 'Replay')}</h2>
              <p className="text-sm text-[#615d59]">
                {t(`参与者：${formatParticipants()}`, `Participants: ${formatParticipants()}`)}
              </p>
              <p className="text-sm text-[#615d59]">
                {t(`状态：${formatStatus(status)}`, `Status: ${formatStatus(status)}`)}
              </p>
              <p className="text-sm text-[#615d59]">
                {t(`共 ${messages.length} 条消息 · ${grouped.length} 轮`, `${messages.length} messages · ${grouped.length} rounds`)}
              </p>
            </div>

            {messages.length === 0 && (
              <div className="text-center py-16 text-[#a39e98]">
                <div className="text-5xl mb-3">📭</div>
                <h3 className="text-lg font-semibold text-[#615d59] mb-1">{t('该讨论暂无消息记录', 'No messages in this discussion')}</h3>
                <button
                  onClick={onBack}
                  className="mt-4 px-4 py-1.5 rounded text-sm font-semibold bg-[#0075de] text-white hover:bg-[#0066cc]"
                >
                  {t('返回列表', 'Back to list')}
                </button>
              </div>
            )}

            {grouped.map((g) => (
              <div key={g.round} className="mb-6">
                <div className="flex items-center gap-3 mb-3">
                  <div className="h-px flex-1 bg-black/[0.06]" />
                  <span className="text-[11px] font-semibold text-[#a39e98] uppercase tracking-wider">
                    {t(`第 ${g.round} 轮`, `Round ${g.round}`)}
                  </span>
                  <div className="h-px flex-1 bg-black/[0.06]" />
                </div>
                <div>
                  {g.messages.map((msg, idx) => (
                    <MessageCard
                      key={msg.id}
                      avatar={msg.avatar}
                      author={msg.author}
                      round={msg.round}
                      content={msg.content}
                      isEven={idx % 2 === 1}
                    />
                  ))}
                </div>
              </div>
            ))}
          </>
        )}
      </main>
    </div>
  );
}
