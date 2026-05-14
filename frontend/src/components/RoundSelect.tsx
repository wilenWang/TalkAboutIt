interface Props {
  value: number;
  onChange: (value: number) => void;
}

export default function RoundSelect({ value, onChange }: Props) {
  return (
    <div className="flex flex-col gap-1" style={{ maxWidth: 120 }}>
      <label className="text-[11px] font-semibold text-[#a39e98] uppercase tracking-wider">
        轮次
      </label>
      <select
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        className="px-2.5 py-2 border border-black/10 rounded text-sm bg-white text-black/95 outline-none focus:border-[#0075de] transition-colors"
      >
        {[1, 2, 3, 4, 5, 10, 15, 20].map((n) => (
          <option key={n} value={n}>
            {n}
          </option>
        ))}
      </select>
    </div>
  );
}
