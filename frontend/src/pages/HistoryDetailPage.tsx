import { useEffect, useState, useMemo } from 'react';
import { getRoundtable } from '../api/client';
import type { PersonaSummary, ReplayMessage } from '../types';
import MessageCard from '../components/MessageCard';

interface Props {
  id: string;
  personaList: PersonaSummary[];
  onBack: () => void;
}

export default function HistoryDetailPage({ id, personaList, onBack }: Props) {
  const [topic, setTopic] = useState('');
  const [messages, setMessages] = useState<ReplayMessage[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    getRoundtable(id)
      .then((snap) => {
        setTopic(snap.topic);
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
        setError(e instanceof Error ? e.message : '加载回放失败');
        setLoading(false);
      });
  }, [id, personaList]);

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

  return (
    <div className="min-h-screen bg-white">
      {/* Header */}
      <header className="px-6 py-3 flex items-center gap-3 border-b border-black/[0.06]">
        <button
          onClick={onBack}
          className="text-sm text-[#615d59] hover:text-black/95 transition-colors"
        >
          ← 返回列表
        </button>
        <span className="text-lg font-bold tracking-tight">✦ TalkAboutIt</span>
        <span className="text-[13px] text-[#a39e98]">回放</span>
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
                  .catch((e) => setError(e instanceof Error ? e.message : '加载回放失败'))
                  .finally(() => setLoading(false));
              }}
              className="px-4 py-1.5 rounded text-sm font-semibold bg-[#0075de] text-white hover:bg-[#0066cc]"
            >
              重试
            </button>
          </div>
        )}

        {/* 回放内容 */}
        {!loading && !error && (
          <>
            <div className="mb-6">
              <h2 className="text-[22px] font-bold tracking-tight mb-1">{topic || '讨论回放'}</h2>
              <p className="text-sm text-[#615d59]">
                共 {messages.length} 条消息 · {grouped.length} 轮
              </p>
            </div>

            {messages.length === 0 && (
              <div className="text-center py-16 text-[#a39e98]">
                <div className="text-5xl mb-3">📭</div>
                <h3 className="text-lg font-semibold text-[#615d59] mb-1">该讨论暂无消息记录</h3>
                <button
                  onClick={onBack}
                  className="mt-4 px-4 py-1.5 rounded text-sm font-semibold bg-[#0075de] text-white hover:bg-[#0066cc]"
                >
                  返回列表
                </button>
              </div>
            )}

            {grouped.map((g) => (
              <div key={g.round} className="mb-6">
                <div className="flex items-center gap-3 mb-3">
                  <div className="h-px flex-1 bg-black/[0.06]" />
                  <span className="text-[11px] font-semibold text-[#a39e98] uppercase tracking-wider">
                    第 {g.round} 轮
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
