import { useEffect, useState } from 'react';
import { listRoundtables } from '../api/client';
import type { RoundtableListItem } from '../api/client';
import type { PersonaSummary } from '../types';

interface Props {
  personaList: PersonaSummary[];
  onSelect: (id: string) => void;
  onBack: () => void;
}

export default function HistoryListPage({ personaList, onSelect, onBack }: Props) {
  const [items, setItems] = useState<RoundtableListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    listRoundtables('completed')
      .then((data) => {
        setItems(data);
        setLoading(false);
      })
      .catch((e) => {
        setError(e instanceof Error ? e.message : '加载失败');
        setLoading(false);
      });
  }, []);

  const formatDate = (s: string) => {
    try {
      const d = new Date(s);
      return d.toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
    } catch {
      return s;
    }
  };

  const getPersonaNames = (ids: string[]) => {
    return ids
      .map((id) => personaList.find((p) => p.id === id)?.name ?? id)
      .join('、');
  };

  return (
    <div className="min-h-screen bg-white">
      {/* Header */}
      <header className="px-6 py-3 flex items-center gap-3 border-b border-black/[0.06]">
        <button
          onClick={onBack}
          className="text-sm text-[#615d59] hover:text-black/95 transition-colors"
        >
          ← 返回首页
        </button>
        <span className="text-lg font-bold tracking-tight">✦ TalkAboutIt</span>
        <span className="text-[13px] text-[#a39e98]">历史记录</span>
      </header>

      <main className="max-w-[720px] mx-auto px-6 py-8">
        <h2 className="text-[22px] font-bold tracking-tight mb-1">讨论历史</h2>
        <p className="text-sm text-[#615d59] mb-6">回顾已完成的圆桌讨论</p>

        {/* 加载态 */}
        {loading && (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="bg-[#f6f5f4] border border-black/[0.06] rounded-lg p-4 animate-pulse">
                <div className="h-4 bg-black/[0.06] rounded w-1/3 mb-2" />
                <div className="h-3 bg-black/[0.04] rounded w-2/3" />
              </div>
            ))}
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
                listRoundtables('completed')
                  .then(setItems)
                  .catch((e) => setError(e instanceof Error ? e.message : '加载失败'))
                  .finally(() => setLoading(false));
              }}
              className="px-4 py-1.5 rounded text-sm font-semibold bg-[#0075de] text-white hover:bg-[#0066cc]"
            >
              重试
            </button>
          </div>
        )}

        {/* 空历史态 */}
        {!loading && !error && items.length === 0 && (
          <div className="text-center py-16 text-[#a39e98]">
            <div className="text-5xl mb-3">📜</div>
            <h3 className="text-lg font-semibold text-[#615d59] mb-1">暂无讨论记录</h3>
            <p className="text-sm mb-4">去开始一场讨论吧</p>
            <button
              onClick={onBack}
              className="px-4 py-2 rounded text-sm font-semibold bg-[#0075de] text-white hover:bg-[#0066cc]"
            >
              开始讨论
            </button>
          </div>
        )}

        {/* 列表 */}
        {!loading && !error && items.length > 0 && (
          <div className="space-y-3">
            {items.map((item) => (
              <div
                key={item.id}
                onClick={() => onSelect(item.id)}
                className="bg-white border border-black/[0.06] rounded-lg p-4 cursor-pointer hover:shadow-[rgba(0,0,0,0.04)_0px_4px_18px] transition-shadow"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="flex-1 min-w-0">
                    <h3 className="text-sm font-semibold text-black/95 truncate mb-1">{item.topic}</h3>
                    <p className="text-[13px] text-[#615d59] truncate">
                      参与者：{getPersonaNames(item.personas)}
                    </p>
                  </div>
                  <div className="flex flex-col items-end gap-1 flex-shrink-0">
                    <span className="bg-[#f2f9ff] text-[#097fe8] text-[11px] font-semibold px-2 py-0.5 rounded-full">
                      {item.max_rounds} 轮
                    </span>
                    <span className="text-[11px] text-[#a39e98]">{formatDate(item.created_at)}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
