import { useMemo } from 'react';
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
  const grouped = useMemo(() => {
    return messages;
  }, [messages]);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="flex items-center gap-2 mb-4">
        <h2 className="text-[22px] font-bold tracking-tight">{t('讨论记录', 'Discussion')}</h2>
        {messages.length > 0 && (
          <span className="bg-[#f2f9ff] text-[#097fe8] text-xs font-semibold px-2.5 py-0.5 rounded-full">
            {t(`${messages.length} 条消息`, `${messages.length} messages`)}
          </span>
        )}
      </div>

      {grouped.length === 0 && !currentSpeaker && (
        <div className="text-center py-16 text-[#a39e98]">
          <div className="text-5xl mb-3">💬</div>
          <h3 className="text-lg font-semibold text-[#615d59] mb-1">{t('暂无消息', 'No messages yet')}</h3>
          <p className="text-sm">{t('讨论尚未开始', 'Discussion has not started yet')}</p>
        </div>
      )}

      {grouped.map((msg, idx) => (
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
        <div className="flex items-center gap-2 px-4 py-3 text-[13px] text-[#a39e98]">
          <span className="text-lg">{currentSpeaker.avatar}</span>
          <span className="font-semibold">{currentSpeaker.name}</span>
          <span>{t('正在输入...', 'Typing...')}</span>
          <span className="flex gap-1 ml-1">
            <span className="w-1 h-1 rounded-full bg-[#a39e98] animate-bounce" style={{ animationDelay: '0ms' }} />
            <span className="w-1 h-1 rounded-full bg-[#a39e98] animate-bounce" style={{ animationDelay: '200ms' }} />
            <span className="w-1 h-1 rounded-full bg-[#a39e98] animate-bounce" style={{ animationDelay: '400ms' }} />
          </span>
        </div>
      )}
    </div>
  );
}
