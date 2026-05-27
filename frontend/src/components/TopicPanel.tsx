import { useState, useRef } from 'react';
import { useLanguage } from '../i18n/LanguageContext';
import RoundSelect from './RoundSelect';

interface Props {
  topic: string;
  onTopicChange: (value: string) => void;
  rounds: number;
  onRoundsChange: (value: number) => void;
  onStart: (topicOverride?: string) => void;
  canStart: boolean;
  loading: boolean;
  hint?: string;
}

export default function TopicPanel({
  topic,
  onTopicChange,
  rounds,
  onRoundsChange,
  onStart,
  canStart,
  loading,
  hint,
}: Props) {
  const { t } = useLanguage();
  const [files, setFiles] = useState<File[]>([]);
  const fileRef = useRef<HTMLInputElement>(null);

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    const dropped = Array.from(e.dataTransfer.files);
    setFiles((prev) => [...prev, ...dropped]);
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      setFiles((prev) => [...prev, ...Array.from(e.target.files!)]);
    }
  };

  const removeFile = (idx: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== idx));
  };

  const readFileAsText = async (file: File): Promise<string> => {
    return new Promise((resolve) => {
      const reader = new FileReader();
      reader.onload = () => resolve(reader.result as string);
      reader.onerror = () => resolve(`[Failed to read: ${file.name}]`);
      reader.readAsText(file);
    });
  };

  const handleStart = async () => {
    let nextTopic = topic;
    if (files.length > 0) {
      const contents = await Promise.all(files.map(readFileAsText));
      const fileSection = contents
        .map((c, i) => `--- ${files[i].name} ---\n${c}`)
        .join('\n\n');
      nextTopic = topic ? `${topic}\n\n${fileSection}` : fileSection;
      onTopicChange(nextTopic);
    }
    onStart(nextTopic);
  };

  return (
    <div className="w-[20%] flex-shrink-0 bg-[#f6f5f4] border-r border-black/[0.06] flex flex-col h-full overflow-hidden">
      <div className="px-4 py-4 border-b border-black/[0.06]">
        <h3 className="text-sm font-bold text-black/95">{t('labelTopic')}</h3>
      </div>

      <div className="flex-1 overflow-y-auto p-4 flex flex-col gap-4">
        {/* Topic textarea */}
        <textarea
          value={topic}
          onChange={(e) => onTopicChange(e.target.value)}
          placeholder={t('inputTopicPlaceholder')}
          rows={6}
          className="w-full px-3 py-2.5 border border-black/10 rounded-lg text-sm bg-white text-black/95 outline-none focus:border-[#0075de] transition-colors resize-none"
        />

        {/* File upload */}
        <div
          onDragOver={(e) => e.preventDefault()}
          onDrop={handleDrop}
          onClick={() => fileRef.current?.click()}
          className="border-2 border-dashed border-black/10 rounded-lg p-4 text-center cursor-pointer hover:border-[#0075de]/40 hover:bg-white/50 transition-colors"
        >
          <input
            ref={fileRef}
            type="file"
            multiple
            className="hidden"
            onChange={handleFileSelect}
          />
          <div className="text-2xl mb-1">📎</div>
          <div className="text-xs text-[#a39e98]">{t('msgUploadFile')}</div>
        </div>

        {/* File list */}
        {files.length > 0 && (
          <div className="flex flex-col gap-1.5">
            {files.map((f, i) => (
              <div
                key={i}
                className="flex items-center gap-2 bg-white rounded-md px-2.5 py-1.5 text-xs border border-black/[0.06]"
              >
                <span className="text-sm">📄</span>
                <span className="flex-1 truncate text-[#615d59]">{f.name}</span>
                <button
                  onClick={(e) => { e.stopPropagation(); removeFile(i); }}
                  className="text-[#a39e98] hover:text-red-500 transition-colors"
                >
                  ✕
                </button>
              </div>
            ))}
          </div>
        )}

        {/* Rounds */}
        <RoundSelect value={rounds} onChange={onRoundsChange} />
      </div>

      {/* Start button */}
      <div className="p-4 border-t border-black/[0.06]">
        <button
          onClick={handleStart}
          disabled={!canStart || loading}
          className={`
            w-full py-2.5 rounded-lg text-sm font-semibold transition-all
            ${!canStart || loading
              ? 'bg-gray-200 text-gray-400 cursor-not-allowed'
              : 'bg-[#0075de] text-white hover:bg-[#0066cc] active:scale-[0.98]'
            }
          `}
        >
          {loading ? t('msgPreparing') : `✦ Talk`}
        </button>
        {hint && (
          <div className="text-[11px] text-[#a39e98] text-center mt-1.5">{hint}</div>
        )}
      </div>
    </div>
  );
}
