import { useLanguage } from '../i18n/LanguageContext';

interface Props {
  avatar: string;
  author: string;
  round: number;
  content: string;
  isEven: boolean;
}

// 简单的 markdown 解析：粗体、斜体、行内代码、代码块、链接、换行
function renderMarkdown(text: string): React.ReactNode[] {
  const lines = text.split('\n');
  const result: React.ReactNode[] = [];
  let inCodeBlock = false;
  let codeContent: string[] = [];
  let codeLang = '';

  const flushCodeBlock = (keyBase: string) => {
    if (!inCodeBlock) return;
    result.push(
      <pre key={`${keyBase}-pre`} className="bg-black/[0.04] rounded-md p-3 overflow-x-auto text-xs my-2">
        {codeLang && <div className="text-[10px] text-[#a39e98] mb-1 uppercase">{codeLang}</div>}
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

    // 代码块边界
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

    // 空行
    if (trimmed === '') {
      result.push(<div key={key} className="h-2" />);
      return;
    }

    // 标题
    if (trimmed.startsWith('### ')) {
      result.push(<h3 key={key} className="text-sm font-bold text-black/95 mt-2 mb-1">{parseInline(trimmed.slice(4))}</h3>);
      return;
    }
    if (trimmed.startsWith('## ')) {
      result.push(<h3 key={key} className="text-sm font-bold text-black/95 mt-2 mb-1">{parseInline(trimmed.slice(3))}</h3>);
      return;
    }
    if (trimmed.startsWith('# ')) {
      result.push(<h3 key={key} className="text-sm font-bold text-black/95 mt-2 mb-1">{parseInline(trimmed.slice(2))}</h3>);
      return;
    }

    // 列表项
    if (trimmed.startsWith('- ') || trimmed.startsWith('* ')) {
      result.push(
        <div key={key} className="flex gap-2 text-sm text-[#615d59] my-0.5">
          <span className="text-[#a39e98] select-none">•</span>
          <span className="flex-1">{parseInline(trimmed.slice(2))}</span>
        </div>
      );
      return;
    }

    // 普通段落
    result.push(<p key={key} className="text-sm text-[#615d59] leading-relaxed my-0.5">{parseInline(line)}</p>);
  });

  if (inCodeBlock) {
    flushCodeBlock('md-end');
  }

  return result;
}

// 允许的链接协议白名单
const ALLOWED_LINK_PROTOCOLS = ['http:', 'https:', 'mailto:'];

function isAllowedLink(href: string): boolean {
  try {
    const url = new URL(href, window.location.href);
    return ALLOWED_LINK_PROTOCOLS.includes(url.protocol);
  } catch {
    // 相对路径或无法解析的链接，默认允许（但当前场景下不应出现）
    return false;
  }
}

function parseInline(text: string): React.ReactNode[] {
  const parts: React.ReactNode[] = [];
  // 按顺序匹配：粗体 **、斜体 *、行内代码 `、链接 [](url)
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
      parts.push(<code key={key++} className="bg-black/[0.05] rounded px-1 py-0.5 text-xs font-mono">{m.slice(1, -1)}</code>);
    } else if (match[2] !== undefined && match[3] !== undefined) {
      const href = match[3];
      if (isAllowedLink(href)) {
        parts.push(
          <a key={key++} href={href} target="_blank" rel="noreferrer" className="text-[#0075de] hover:underline">
            {match[2]}
          </a>
        );
      } else {
        // 危险协议降级为纯文本
        parts.push(<span key={key++}>{match[2]} ({href})</span>);
      }
    } else {
      parts.push(<span key={key++}>{m}</span>);
    }
    lastIndex = regex.lastIndex;
  }

  if (lastIndex < text.length) {
    parts.push(<span key={key++}>{text.slice(lastIndex)}</span>);
  }

  return parts;
}

export default function MessageCard({ avatar, author, round, content, isEven }: Props) {
  const { f } = useLanguage();
  return (
    <div
      className={`
        border border-black/[0.06] rounded-lg px-5 py-4 mb-2 shadow-[0_1px_3px_rgba(0,0,0,0.02)]
        ${isEven ? 'bg-[#f6f5f4]' : 'bg-white'}
      `}
    >
      <div className="flex items-center gap-2 mb-1.5">
        <span className="text-lg leading-none">{avatar}</span>
        <span className="text-[13px] font-semibold">{author}</span>
        <span className="text-[11px] text-[#a39e98]">{f('fmtRoundLabel', { n: round })}</span>
      </div>
      <div className="text-sm leading-relaxed text-[#615d59]">{renderMarkdown(content)}</div>
    </div>
  );
}
