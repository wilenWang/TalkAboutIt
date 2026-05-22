import { useEffect, useRef } from 'react';
import MessageCard from './MessageCard';
import { useLanguage } from '../i18n/LanguageContext';

export interface StreamMessage {
  id: string;
  avatar: string;
  author: string;
  personaId: string;
  round: number;
  content: string;
  status: 'streaming' | 'done';
}

interface Props {
  messages: StreamMessage[];
  currentSpeaker: { name: string; avatar: string } | null;
}

export default function MessageStream({ messages, currentSpeaker }: Props) {
  const { t } = useLanguage();
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, currentSpeaker]);

  return (
    <div className="flex-1 flex flex-col h-full overflow-hidden">
      {messages.length === 0 && !currentSpeaker ? (
        <div className="flex-1 flex flex-col items-center justify-center text-[#a39e98]">
          <div className="text-5xl mb-3">💬</div>
          <h3 className="text-base font-semibold text-[#615d59] mb-1">{t('msgNoMessages')}</h3>
          <p className="text-sm">{t('msgNotStarted')}</p>
        </div>
      ) : (
        <div className="flex-1 overflow-y-auto px-5 py-4">
          {messages.map((msg, idx) => (
            <MessageCard
              key={msg.id}
              avatar={msg.avatar}
              author={msg.author}
              round={msg.round}
              content={msg.content}
              isEven={idx % 2 === 1}
            />
          ))}

          {currentSpeaker && (
            <div className="flex items-center gap-2.5 mb-3">
              <div className="w-9 h-9 rounded-lg bg-white border border-black/[0.06] flex items-center justify-center text-lg">
                {currentSpeaker.avatar}
              </div>
              <div className="flex items-center gap-1.5">
                <span className="text-[13px] font-semibold text-black/90">{currentSpeaker.name}</span>
                <span className="text-[12px] text-[#a39e98]">{t('labelTyping')}</span>
                <span className="flex gap-0.5">
                  <span className="w-1 h-1 rounded-full bg-[#a39e98] animate-bounce" style={{ animationDelay: '0ms' }} />
                  <span className="w-1 h-1 rounded-full bg-[#a39e98] animate-bounce" style={{ animationDelay: '200ms' }} />
                  <span className="w-1 h-1 rounded-full bg-[#a39e98] animate-bounce" style={{ animationDelay: '400ms' }} />
                </span>
              </div>
            </div>
          )}
          <div ref={bottomRef} />
        </div>
      )}
    </div>
  );
}
