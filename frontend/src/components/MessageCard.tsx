import { useLanguage } from '../i18n/LanguageContext';
import { isImageAvatar } from '../utils/avatar';

interface Props {
  avatar: string;
  author: string;
  round: number;
  content: string;
  isEven: boolean;
}

function renderMarkdown(text: string): React.ReactNode[] {
  const lines = text.split('\n');
  const result: React.ReactNode[] = [];
  let inCodeBlock = false;
  let codeContent: string[] = [];
  let codeLang = '';

  const flushCodeBlock = (keyBase: string) => {
    if (!inCodeBlock) return;
    result.push(
      <pre key={`${keyBase}-pre`} className="bg-black/[0.06] rounded p-2 overflow-x-auto text-[11px] my-1">
        {codeLang && <div className="text-[9px] text-[#a39e98] mb-0.5 uppercase">{codeLang}</div>}
        <code className="text-[#615d59] whitespace-pre">{codeContent.join('\n')}</code>
      </pre>
    );
    inCodeBlock = false;
    codeContent = [];
    codeLang = '';
  };

  lines.forEach((line, idx) => {
    const key = `md-${idx}`;
    const trimmed = line.trim();

    if (trimmed.startsWith('```')) {
      if (inCodeBlock) {
        flushCodeBlock(key);
      } else {
        inCodeBlock = true;
        codeLang = trimmed.slice(3).trim();
      }
      return;
    }

    if (inCodeBlock) {
      codeContent.push(line);
      return;
    }

    if (trimmed === '') {
      result.push(<div key={key} className="h-1" />);
      return;
    }

    if (trimmed.startsWith('### ')) {
      result.push(<div key={key} className="text-xs font-bold text-black/90 mt-1.5 mb-0.5">{parseInline(trimmed.slice(4))}</div>);
      return;
    }
    if (trimmed.startsWith('## ')) {
      result.push(<div key={key} className="text-xs font-bold text-black/90 mt-1.5 mb-0.5">{parseInline(trimmed.slice(3))}</div>);
      return;
    }
    if (trimmed.startsWith('# ')) {
      result.push(<div key={key} className="text-xs font-bold text-black/90 mt-1.5 mb-0.5">{parseInline(trimmed.slice(2))}</div>);
      return;
    }

    if (trimmed.startsWith('- ') || trimmed.startsWith('* ')) {
      result.push(
        <div key={key} className="flex gap-1.5 text-[13px] text-[#615d59] my-0.5">
          <span className="text-[#a39e98] select-none">•</span>
          <span className="flex-1">{parseInline(trimmed.slice(2))}</span>
        </div>
      );
      return;
    }

    result.push(<p key={key} className="text-[13px] text-[#615d59] leading-relaxed my-0.5">{parseInline(line)}</p>);
  });

  if (inCodeBlock) {
    flushCodeBlock('md-end');
  }

  return result;
}

const ALLOWED_LINK_PROTOCOLS = ['http:', 'https:', 'mailto:'];

function isAllowedLink(href: string): boolean {
  try {
    const url = new URL(href, window.location.href);
    return ALLOWED_LINK_PROTOCOLS.includes(url.protocol);
  } catch {
    return false;
  }
}

function parseInline(text: string): React.ReactNode[] {
  const parts: React.ReactNode[] = [];
  const regex = /(\*\*[\s\S]*?\*\*|\*[\s\S]*?\*|`[^`]+`|\[([^\]]+)\]\(([^)]+)\))/g;
  let lastIndex = 0;
  let match: RegExpExecArray | null;
  let key = 0;

  while ((match = regex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      parts.push(<span key={key++}>{text.slice(lastIndex, match.index)}</span>);
    }
    const m = match[0];
    if (m.startsWith('**') && m.endsWith('**')) {
      parts.push(<strong key={key++} className="font-semibold text-black/95">{m.slice(2, -2)}</strong>);
    } else if (m.startsWith('*') && m.endsWith('*')) {
      parts.push(<em key={key++} className="italic">{m.slice(1, -1)}</em>);
    } else if (m.startsWith('`') && m.endsWith('`')) {
      parts.push(<code key={key++} className="bg-black/[0.05] rounded px-1 py-0.5 text-[11px] font-mono">{m.slice(1, -1)}</code>);
    } else if (match[2] !== undefined && match[3] !== undefined) {
      const href = match[3];
      if (isAllowedLink(href)) {
        parts.push(<a key={key++} href={href} target="_blank" rel="noreferrer" className="text-[#0075de] hover:underline">{match[2]}</a>);
      } else {
        parts.push(<span key={key++}>{match[2]} ({href})</span>);
      }
    } else {
      parts.push(<span key={key++}>{m}</span>);
    }
    lastIndex = regex.lastIndex;
  }

  if (lastIndex < text.length) {
    parts.push(<span key={key}>{text.slice(lastIndex)}</span>);
  }

  return parts;
}

export default function MessageCard({ avatar, author, round, content, isEven }: Props) {
  const { f } = useLanguage();

  return (
    <div className="flex gap-2.5 mb-3">
      {/* Avatar */}
      <div className="flex-shrink-0">
        {isImageAvatar(avatar) ? (
          <img src={avatar} alt={author} className="w-9 h-9 rounded-lg object-cover" />
        ) : (
          <div className="w-9 h-9 rounded-lg bg-white border border-black/[0.06] flex items-center justify-center text-lg">
            {avatar}
          </div>
        )}
      </div>

      {/* Bubble */}
      <div className="flex-1 min-w-0">
        <div className="flex items-baseline gap-2 mb-1">
          <span className="text-[13px] font-semibold text-black/90">{author}</span>
          <span className="text-[10px] text-[#a39e98]">{f('fmtRoundLabel', { n: round })}</span>
        </div>
        <div className={`inline-block rounded-lg px-3.5 py-2.5 max-w-full ${isEven ? 'bg-[#f0f0f0]' : 'bg-white border border-black/[0.06]'}`}>
          <div className="break-words">{renderMarkdown(content)}</div>
        </div>
      </div>
    </div>
  );
}
