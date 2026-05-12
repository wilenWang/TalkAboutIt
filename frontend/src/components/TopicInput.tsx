interface Props {
  value: string;
  onChange: (value: string) => void;
}

export default function TopicInput({ value, onChange }: Props) {
  return (
    <div className="flex-1 flex flex-col gap-1">
      <label className="text-[11px] font-semibold text-[#a39e98] uppercase tracking-wider">
        讨论话题
      </label>
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="输入讨论话题..."
        className="px-2.5 py-2 border border-black/10 rounded text-sm bg-white text-black/95 outline-none focus:border-[#0075de] transition-colors"
      />
    </div>
  );
}
