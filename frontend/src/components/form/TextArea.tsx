interface Props {
  label: string;
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  rows?: number;
}

export default function TextArea({ label, value, onChange, placeholder, rows = 3 }: Props) {
  return (
    <div>
      <label className="block text-xs font-medium text-[#615d59] mb-1">{label}</label>
      <textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        rows={rows}
        className="w-full px-2.5 py-2 border border-black/10 rounded text-sm bg-white text-black/95 outline-none focus:border-[#0075de] transition-colors resize-none"
      />
    </div>
  );
}
