interface Props {
  label: string;
  value: number;
  onChange: (v: number) => void;
  min: number;
  max: number;
}

export default function NumberInput({ label, value, onChange, min, max }: Props) {
  return (
    <div>
      <label className="block text-xs font-medium text-[#615d59] mb-1">{label}</label>
      <input
        type="number"
        min={min}
        max={max}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        className="w-full px-2.5 py-2 border border-black/10 rounded text-sm bg-white text-black/95 outline-none focus:border-[#0075de] transition-colors"
      />
    </div>
  );
}
