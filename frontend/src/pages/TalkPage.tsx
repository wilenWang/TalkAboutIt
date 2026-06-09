import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import TopicPanel from '../components/TopicPanel';
import PersonaSelector from '../components/PersonaSelector';
import MessageStream from '../components/MessageStream';
import { createRoundtable, startRoundtable, getRoundtable } from '../api/client';
import { useSSE } from '../hooks/useSSE';
import { useAppStore } from '../stores/useAppStore';
import { usePersonaStore } from '../stores/usePersonaStore';
import { useLanguage } from '../i18n/LanguageContext';
import type { PersonaSummary } from '../types';

export default function TalkPage() {
  const navigate = useNavigate();
  const { language, t } = useLanguage();
  const { personas: personaList, fetchPersonas } = usePersonaStore();

  const {
    topic,
    rounds,
    selectedPersonas,
    status,
    messages,
    currentSpeaker,
    error,
    rtId,
    setTopic,
    setRounds,
    startCreating,
    startStreaming,
    complete,
    setCurrentSpeaker,
    setError,
    setRtId,
    setCurrentRound,
    addMessageChunk,
    finalizeMessage,
    abortMessage,
    setMessages,
    setSelectedPersonas,
  } = useAppStore();

  const [resumeFromEventId, setResumeFromEventId] = useState('');
  const [sseUrl, setSseUrl] = useState<string | null>(null);

  const handleSSEMessage = useCallback((msg: import('../hooks/useSSE').SSEMessage) => {
    switch (msg.event) {
      case 'stream_start':
        startStreaming();
        setMessages([]);
        break;

      case 'speaking': {
        const data = msg.data as { persona_name: string; avatar: string };
        setCurrentSpeaker({ name: data.persona_name, avatar: data.avatar });
        break;
      }

      case 'message_chunk': {
        const data = msg.data as { chunk: string; persona_id: string; round: number; persona_name?: string; avatar?: string };
        addMessageChunk(data);
        break;
      }

      case 'message_done': {
        const data = msg.data as {
          message_id: string;
          persona_id: string;
          persona_name: string;
          avatar: string;
          round: number;
          content: string;
        };
        finalizeMessage(data);
        break;
      }

      case 'message_aborted': {
        const data = msg.data as { persona_id: string; round: number };
        abortMessage(data.persona_id);
        break;
      }

      case 'round_start': {
        const data = msg.data as { round: number };
        setCurrentRound(data.round);
        break;
      }

      case 'stream_done':
        complete();
        setCurrentSpeaker(null);
        break;

      case 'error': {
        const data = msg.data as { error: string; recoverable?: boolean };
        setError(data.error);
        if (!data.recoverable) {
          // Reset status to idle on non-recoverable error
          useAppStore.setState({ status: 'idle' });
        }
        break;
      }
    }
  }, [startStreaming, setMessages, setCurrentSpeaker, addMessageChunk, finalizeMessage, abortMessage, setCurrentRound, complete, setError]);

  const { status: sseStatus, reconnect: reconnectSSE } = useSSE(
    sseUrl,
    handleSSEMessage,
    (err) => {
      console.error('SSE error:', err);
    },
    { initialLastEventId: resumeFromEventId || undefined }
  );

  const loadSnapshot = useCallback(async (id: string) => {
    try {
      const snap = await getRoundtable(id);
      setRtId(snap.id);
      setTopic(snap.topic);
      setRounds(snap.max_rounds);
      setSelectedPersonas(snap.personas);

      const historyMsgs = snap.messages.map((m) => {
        const matched = personaList.find((p) => p.id === m.persona_id);
        const authorLabel = language === 'zh-CN'
          ? (matched?.display_name ?? matched?.name ?? m.persona_id)
          : (matched?.name ?? m.persona_id);
        return {
          id: m.id,
          avatar: matched?.avatar ?? '🤖',
          author: authorLabel,
          personaId: m.persona_id,
          round: m.round,
          content: m.content,
          status: 'done' as const,
        };
      });
      setMessages(historyMsgs);

      if (snap.status === 'completed') {
        useAppStore.setState({ status: 'completed' });
      } else if (snap.status === 'running') {
        startStreaming();
        setResumeFromEventId(String(snap.last_event_id));
        setSseUrl(`/api/v1/roundtables/${id}/events`);
      } else {
        useAppStore.setState({ status: 'idle' });
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : t('errSnapshotFailed'));
    }
  }, [
    personaList,
    setError,
    setMessages,
    setRounds,
    setRtId,
    setSelectedPersonas,
    setTopic,
    startStreaming,
    t,
  ]);

  // Load persona list on mount
  useEffect(() => {
    fetchPersonas();
  }, [fetchPersonas]);

  // Check URL param ?rt={id} on mount
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const rtId = params.get('rt');
    if (rtId) {
      loadSnapshot(rtId);
    }
  }, [loadSnapshot]);

  // Update messages when personaList loads (for avatar/name matching)
  useEffect(() => {
    if (personaList.length > 0 && messages.length > 0) {
      const updated = messages.map((m) => {
        if (m.status === 'done') {
          const matched = personaList.find((p: PersonaSummary) => p.id === m.personaId);
          if (matched) {
            const expectedAuthor = language === 'zh-CN'
              ? (matched.display_name ?? matched.name)
              : matched.name;
            if (m.author !== expectedAuthor || m.avatar !== matched.avatar) {
              return { ...m, author: expectedAuthor, avatar: matched.avatar };
            }
          }
        }
        return m;
      });
      const changed = updated.some((m, i) => m !== messages[i]);
      if (changed) {
        setMessages(updated);
      }
    }
  }, [personaList, messages, setMessages]);

  const handleStart = async (topicOverride?: string) => {
    if (selectedPersonas.length < 2) return;
    setError(null);
    startCreating();

    try {
      const finalTopic = topicOverride ?? topic;
      const rt = await createRoundtable({
        topic: finalTopic,
        personas: selectedPersonas,
        max_rounds: rounds,
        language,
      });

      setRtId(rt.id);
      window.history.pushState({}, '', `/?rt=${rt.id}`);
      setResumeFromEventId('');
      setSseUrl(`/api/v1/roundtables/${rt.id}/events`);

      await startRoundtable(rt.id);
    } catch (e) {
      setError(e instanceof Error ? e.message : t('errUnknown'));
      useAppStore.setState({ status: 'idle' });
    }
  };

  const canStart = selectedPersonas.length >= 2 && topic.trim().length > 0 && status === 'idle';
  const startHint = selectedPersonas.length < 2 ? t('msgMinParticipants') : undefined;

  return (
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
      <PersonaSelector />

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
        <MessageStream messages={messages} currentSpeaker={currentSpeaker} />

        {/* Completed: view replay */}
        {status === 'completed' && rtId && (
          <div className="px-4 py-3 border-t border-black/[0.06] flex justify-center">
            <button
              onClick={() => {
                navigate(`/history/${rtId}`);
              }}
              className="px-5 py-2 rounded-lg text-sm font-semibold bg-[#0075de] text-white hover:bg-[#0066cc]"
            >
              {t('actionViewReplay')}
            </button>
          </div>
        )}
      </div>
    </>
  );
}
