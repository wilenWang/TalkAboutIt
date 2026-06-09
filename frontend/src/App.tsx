import { useState, useCallback, useRef, useEffect } from 'react';
import PersonaSelector from './components/PersonaSelector';
import NavSidebar from './components/NavSidebar';
import TopicPanel from './components/TopicPanel';
import MessageStream, { StreamMessage } from './components/MessageStream';
import { createRoundtable, startRoundtable, getRoundtable, fetchPersonas } from './api/client';
import { useSSE } from './hooks/useSSE';
import type { PersonaSummary } from './types';
import HistoryListPage from './pages/HistoryListPage';
import HistoryDetailPage from './pages/HistoryDetailPage';
import PersonaManagePage from './pages/PersonaManagePage';
import { useLanguage, type SystemLanguage } from './i18n/LanguageContext';

type AppStatus = 'idle' | 'creating' | 'streaming' | 'completed';
type Page = 'talk' | 'history' | 'history-detail' | 'personas';

// 从 URL pathname 解析初始页面和 historyId
function parsePageFromPath(): { page: Page; historyId: string | null } {
  const path = window.location.pathname;
  if (path === '/history') {
    return { page: 'history', historyId: null };
  }
  if (path.startsWith('/history/')) {
    const id = path.slice('/history/'.length);
    if (id) {
      return { page: 'history-detail', historyId: id };
    }
  }
  if (path === '/personas') {
    return { page: 'personas', historyId: null };
  }
  return { page: 'talk', historyId: null };
}

export default function App() {
  const { language: systemLanguage, setLanguage: setSystemLanguage, t } = useLanguage();
  const initial = parsePageFromPath();
  const [page, setPage] = useState<Page>(initial.page);
  const [historyId, setHistoryId] = useState<string | null>(initial.historyId);

  const [selectedPersonas, setSelectedPersonas] = useState<string[]>([]);
  const [topic, setTopic] = useState(
    systemLanguage === 'en-US' ? 'Will AI replace programmers?' : 'AI 会取代程序员吗？'
  );
  const [rounds, setRounds] = useState(3);
  // 讨论语言统一走右上角全局语言切换，不再单独设置
  const debateLanguage = systemLanguage;
  const [status, setStatus] = useState<AppStatus>('idle');
  const [messages, setMessages] = useState<StreamMessage[]>([]);
  const [currentSpeaker, setCurrentSpeaker] = useState<{ name: string; nameEn: string; avatar: string } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [languageMenuOpen, setLanguageMenuOpen] = useState(false);
  // 本地 persona 列表，用于快照恢复时匹配作者信息
  const [personaList, setPersonaList] = useState<PersonaSummary[]>([]);

  const messagesRef = useRef<StreamMessage[]>([]);
  const currentRoundRef = useRef(1);
  const rtIdRef = useRef<string | null>(null);
  const languageMenuRef = useRef<HTMLDivElement | null>(null);
  // 快照恢复后用于 SSE 首次连接的 last_event_id
  const resumeFromEventIdRef = useRef<string>('');



  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (languageMenuRef.current && !languageMenuRef.current.contains(event.target as Node)) {
        setLanguageMenuOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // 加载本地 persona 列表
  useEffect(() => {
    fetchPersonas()
      .then(setPersonaList)
      .catch(console.error);
  }, []);

  const handleSSEMessage = useCallback((msg: import('./hooks/useSSE').SSEMessage) => {
    switch (msg.event) {
      case 'stream_start':
        setStatus('streaming');
        setMessages([]);
        messagesRef.current = [];
        break;

      case 'speaking': {
        const data = msg.data as { persona_name: string; persona_name_en: string; avatar: string };
        setCurrentSpeaker({ name: data.persona_name, nameEn: data.persona_name_en, avatar: data.avatar });
        break;
      }

      case 'message_chunk': {
        // 实时更新当前发言内容（追加到当前 speaker 的临时消息）
        const data = msg.data as { chunk: string; persona_id: string; round: number };
        setMessages((prev) => {
          const last = prev[prev.length - 1];
          if (last && last.status === 'streaming' && last.personaId === data.persona_id) {
            const updated = [...prev];
            updated[updated.length - 1] = {
              ...last,
              content: last.content + data.chunk,
            };
            messagesRef.current = updated;
            return updated;
          }
          // 如果没有正在流式传输的消息，创建一个新的
          // 使用 msg.data 中可能携带的 persona_name/avatar，否则从 currentSpeaker 取
          const speakerName = (msg.data as Record<string, unknown>).persona_name as string | undefined;
          const speakerNameEn = (msg.data as Record<string, unknown>).persona_name_en as string | undefined;
          const speakerAvatar = (msg.data as Record<string, unknown>).avatar as string | undefined;
          const authorName = systemLanguage === 'zh-CN'
            ? (speakerName ?? currentSpeakerRef.current?.name ?? 'AI')
            : (speakerNameEn ?? currentSpeakerRef.current?.nameEn ?? speakerName ?? currentSpeakerRef.current?.name ?? 'AI');
          const newMsg: StreamMessage = {
            id: `temp_${data.persona_id}_${Date.now()}`,
            avatar: speakerAvatar ?? currentSpeakerRef.current?.avatar ?? '🤖',
            author: authorName,
            personaId: data.persona_id,
            round: data.round,
            content: data.chunk,
            status: 'streaming',
          };
          const updated = [...prev, newMsg];
          messagesRef.current = updated;
          return updated;
        });
        break;
      }

      case 'message_done': {
        const data = msg.data as {
          message_id: string;
          persona_id: string;
          persona_name: string;
          persona_name_en: string;
          avatar: string;
          round: number;
          content: string;
        };
        const newMsg: StreamMessage = {
          id: data.message_id,
          avatar: data.avatar,
          author: systemLanguage === 'zh-CN' ? data.persona_name : data.persona_name_en,
          personaId: data.persona_id,
          round: data.round,
          content: data.content,
          status: 'done',
        };
        messagesRef.current = [...messagesRef.current.filter(m => m.id !== newMsg.id && !(m.status === 'streaming' && m.personaId === data.persona_id)), newMsg];
        setMessages(messagesRef.current);
        setCurrentSpeaker(null);
        break;
      }

      case 'message_aborted': {
        // 流式发言被可恢复错误中断，清理对应临时消息并取消当前 speaker
        const data = msg.data as { persona_id: string; round: number };
        setMessages((prev) => {
          const updated = prev.filter(m => !(m.status === 'streaming' && m.personaId === data.persona_id));
          messagesRef.current = updated;
          return updated;
        });
        setCurrentSpeaker(null);
        break;
      }

      case 'round_start': {
        const data = msg.data as { round: number };
        currentRoundRef.current = data.round;
        break;
      }

      case 'stream_done':
        setStatus('completed');
        setCurrentSpeaker(null);
        break;

      case 'error': {
        const data = msg.data as { error: string; recoverable?: boolean };
        setError(data.error);
        if (!data.recoverable) {
          setStatus('idle');
        }
        break;
      }
    }
  }, []);

  // currentSpeaker 的 ref，供 message_chunk 不依赖 currentSpeaker state 时使用
  const currentSpeakerRef = useRef(currentSpeaker);
  useEffect(() => {
    currentSpeakerRef.current = currentSpeaker;
  }, [currentSpeaker]);

  const [sseUrl, setSseUrl] = useState<string | null>(null);

  const { status: sseStatus, reconnect: reconnectSSE } = useSSE(
    sseUrl,
    handleSSEMessage,
    (err) => {
      console.error('SSE error:', err);
    },
    { initialLastEventId: resumeFromEventIdRef.current || undefined }
  );

  // 页面加载时检查 URL 参数 ?rt={id}，加载快照
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const rtId = params.get('rt');
    if (rtId) {
      loadSnapshot(rtId);
    }
  }, []);

  // 同步 page/historyId 到 URL（pushState/replaceState）
  useEffect(() => {
    let url: string;
    if (page === 'history') {
      url = '/history';
    } else if (page === 'history-detail' && historyId) {
      url = `/history/${historyId}`;
    } else if (page === 'personas') {
      url = '/personas';
    } else {
      url = '/';
    }
    // 仅在 URL 不同时更新，避免无限循环
    if (window.location.pathname !== url) {
      window.history.pushState({ page, historyId }, '', url);
    }
  }, [page, historyId]);

  // 监听浏览器前进后退（popstate），从 URL 恢复 page/historyId
  useEffect(() => {
    const handlePopState = () => {
      const next = parsePageFromPath();
      setPage(next.page);
      setHistoryId(next.historyId);
    };
    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, []);

  // 加载 roundtable 快照
  const loadSnapshot = async (id: string) => {
    try {
      const snap = await getRoundtable(id);
      rtIdRef.current = snap.id;
      setTopic(snap.topic);
      setRounds(snap.max_rounds);
      setSelectedPersonas(snap.personas);
      // 讨论语言统一走全局语言，快照中的语言仅用于展示兼容，不再单独设置

      // 将历史消息转为 StreamMessage，并尝试从本地 personaList 匹配真实名字和头像
      const historyMsgs: StreamMessage[] = snap.messages.map((m) => {
        const matched = personaList.find(p => p.id === m.persona_id);
        const authorLabel = systemLanguage === 'zh-CN'
          ? (matched?.display_name ?? matched?.name ?? m.persona_id)
          : (matched?.name ?? m.persona_id);
        return {
          id: m.id,
          avatar: matched?.avatar ?? '🤖',
          author: authorLabel,
          personaId: m.persona_id,
          round: m.round,
          content: m.content,
          status: 'done',
        };
      });
      messagesRef.current = historyMsgs;
      setMessages(historyMsgs);

      if (snap.status === 'completed') {
        setStatus('completed');
      } else if (snap.status === 'running') {
        setStatus('streaming');
        // 恢复快照时带上 last_event_id，使 SSE 首次连接即可从正确位置恢复
        resumeFromEventIdRef.current = String(snap.last_event_id);
        setSseUrl(`/api/v1/roundtables/${id}/events`);
      } else {
        setStatus('idle');
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : t('errSnapshotFailed'));
    }
  };

  // personaList 加载完成后再尝试刷新一次快照中的作者信息
  useEffect(() => {
    if (personaList.length > 0 && messages.length > 0) {
      setMessages((prev) => {
        let changed = false;
        const updated = prev.map((m) => {
          if (m.status === 'done') {
            const matched = personaList.find(p => p.id === m.personaId);
            if (matched) {
              const expectedAuthor = systemLanguage === 'zh-CN'
                ? (matched.display_name ?? matched.name)
                : matched.name;
              if (m.author !== expectedAuthor || m.avatar !== matched.avatar) {
                changed = true;
                return { ...m, author: expectedAuthor, avatar: matched.avatar };
              }
            }
          }
          return m;
        });
        if (changed) {
          messagesRef.current = updated;
        }
        return changed ? updated : prev;
      });
    }
  }, [personaList]);

  const handleStart = async () => {
    if (selectedPersonas.length < 2) return;
    setError(null);
    setStatus('creating');

    try {
      const rt = await createRoundtable({
        topic,
        personas: selectedPersonas,
        max_rounds: rounds,
        language: debateLanguage,
      });

      rtIdRef.current = rt.id;
      // 更新 URL，支持刷新恢复（SSE 连接参数保留为 search 参数）
      window.history.pushState({ page: 'talk', historyId: null }, '', `/?rt=${rt.id}`);

      resumeFromEventIdRef.current = '';
      setSseUrl(`/api/v1/roundtables/${rt.id}/events`);

      await startRoundtable(rt.id);
    } catch (e) {
      setError(e instanceof Error ? e.message : t('errUnknown'));
      setStatus('idle');
    }
  };

  const canStart = selectedPersonas.length >= 2 && topic.trim().length > 0 && status === 'idle';
  const startHint = selectedPersonas.length < 2 ? t('msgMinParticipants') : undefined;

  // 连接状态提示文案
  const connectionBadge = () => {
    if (sseStatus === 'reconnecting') {
      return <span className="text-xs text-orange-500 font-medium animate-pulse">↻ {t('statusReconnecting')}</span>;
    }
    if (sseStatus === 'connecting') {
      return <span className="text-xs text-[#0075de] font-medium animate-pulse">● {t('statusConnecting')}</span>;
    }
    if (sseStatus === 'disconnected') {
      return <span className="text-xs text-red-500 font-medium">✗ {t('statusDisconnected')}</span>;
    }
    if (status === 'streaming') {
      return <span className="text-xs text-[#0075de] font-medium animate-pulse">● {t('statusInProgress')}</span>;
    }
    if (status === 'completed') {
      return <span className="text-xs text-green-600 font-medium">✓ {t('statusCompleted')}</span>;
    }
    return null;
  };

  const systemLanguageOptions: { value: SystemLanguage; label: string; shortLabel: string; flag: string }[] = [
    { value: 'zh-CN', label: '简体中文', shortLabel: '简体中文', flag: '🇨🇳' },
    { value: 'en-US', label: 'English', shortLabel: 'English', flag: '🇺🇸' },
  ];

  const currentSystemLanguage = systemLanguageOptions.find((option) => option.value === systemLanguage)!;

  // 页面路由
  if (page === 'history') {
    return (
      <HistoryListPage
        personaList={personaList}
        onSelect={(id) => {
          setHistoryId(id);
          setPage('history-detail');
        }}
        onBack={() => setPage('talk')}
      />
    );
  }

  if (page === 'history-detail' && historyId) {
    return (
      <HistoryDetailPage
        id={historyId}
        personaList={personaList}
        onBack={() => setPage('history')}
      />
    );
  }

  // talk 和 personas 页面共享侧边栏布局
  const navPage = page === 'personas' ? 'personas' : 'talk';

  return (
    <div className="h-screen flex flex-col">
      {/* Header */}
      <header className="px-6 py-3 flex items-center gap-3 border-b border-black/[0.06]">
        <span className="text-lg font-bold tracking-tight">✦ TalkAboutIt</span>
        <span className="text-[13px] text-[#a39e98]">{navPage === 'talk' ? t('pageRoundtable') : t('tabPersonas')}</span>
        <span className="flex-1" />
        <div className="relative" ref={languageMenuRef}>
          <button
            type="button"
            onClick={() => setLanguageMenuOpen((open) => !open)}
            className="text-[13px] text-[#615d59] hover:text-black/95 transition-colors"
          >
            🌐 {currentSystemLanguage.shortLabel}
          </button>
          {languageMenuOpen && (
            <div className="absolute right-0 top-full mt-2 min-w-[168px] rounded-xl bg-white shadow-[0_10px_30px_rgba(0,0,0,0.08)] border border-black/[0.06] py-1 z-20">
              {systemLanguageOptions.map((option) => {
                const isActive = option.value === systemLanguage;
                return (
                  <button
                    key={option.value}
                    type="button"
                    onClick={() => {
                      setSystemLanguage(option.value);
                      setLanguageMenuOpen(false);
                    }}
                    className={`w-full px-3 py-2 text-left text-[13px] transition-colors ${
                      isActive ? 'bg-[#f2f9ff] text-[#0075de] font-semibold' : 'text-[#615d59] hover:bg-black/[0.03]'
                    }`}
                  >
                    {option.flag} {option.label}
                  </button>
                );
              })}
            </div>
          )}
        </div>
        <button
          onClick={() => setPage('history')}
          className="text-[13px] text-[#615d59] hover:text-black/95 transition-colors"
        >
          {t('pageHistory')}
        </button>
        {connectionBadge()}
      </header>

      {/* Main */}
      <div className="flex flex-1 overflow-hidden">
        <NavSidebar
          active={navPage}
          onNavigate={(p) => setPage(p === 'talk' ? 'talk' : 'personas')}
        />

        {navPage === 'personas' ? (
          <main className="flex-1 overflow-y-auto">
            <PersonaManagePage onBack={() => setPage('talk')} />
          </main>
        ) : (
          <>
            {/* Column 1: Topic setup */}
            <TopicPanel
              topic={topic}
              onTopicChange={setTopic}
              rounds={rounds}
              onRoundsChange={setRounds}
              onStart={handleStart}
              canStart={canStart}
              loading={status === 'creating'}
              hint={startHint}
            />

            {/* Column 2: Persona selection */}
            <PersonaSelector selected={selectedPersonas} onChange={setSelectedPersonas} />

            {/* Column 3: Chat area */}
            <div className="flex-1 flex flex-col overflow-hidden bg-[#f9f9f9]">
              {/* Error banner */}
              {error && (
                <div className="px-4 py-2 bg-red-50 border-b border-red-100 flex items-center gap-2">
                  <span className="text-sm text-red-500">{error}</span>
                  <button
                    onClick={() => {
                      setError(null);
                      if (status === 'idle') handleStart();
                    }}
                    className="text-sm text-[#0075de] hover:underline"
                  >
                    {t('actionRetry')}
                  </button>
                </div>
              )}

              {/* Disconnection banner */}
              {sseStatus === 'disconnected' && status === 'streaming' && (
                <div className="px-4 py-2 bg-red-50 border-b border-red-100 flex items-center justify-between">
                  <span className="text-sm text-red-600">{t('statusConnectionLost')}</span>
                  <button
                    onClick={() => reconnectSSE()}
                    className="text-sm font-semibold text-red-600 hover:text-red-700"
                  >
                    {t('statusReconnectManually')}
                  </button>
                </div>
              )}

              {/* Messages */}
              <MessageStream messages={messages} currentSpeaker={
                currentSpeaker ? {
                  name: systemLanguage === 'zh-CN' ? currentSpeaker.name : currentSpeaker.nameEn,
                  avatar: currentSpeaker.avatar
                } : null
              } />

              {/* Completed: view replay */}
              {status === 'completed' && rtIdRef.current && (
                <div className="px-4 py-3 border-t border-black/[0.06] flex justify-center">
                  <button
                    onClick={() => {
                      const id = rtIdRef.current;
                      if (id) {
                        setHistoryId(id);
                        setPage('history-detail');
                        window.history.pushState({ page: 'history-detail', historyId: id }, '', `/history/${id}`);
                      }
                    }}
                    className="px-5 py-2 rounded-lg text-sm font-semibold bg-[#0075de] text-white hover:bg-[#0066cc]"
                  >
                    {t('actionViewReplay')}
                  </button>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}