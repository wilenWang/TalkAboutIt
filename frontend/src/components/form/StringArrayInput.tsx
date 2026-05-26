import { useState, useEffect } from 'react';

interface Props {
  label: string;
  values: string[];
  onChange: (v: string[]) => void;
}

export default function StringArrayInput({ label, values, onChange }: Props) {
  const [text, setText] = useState(values.join('\n'));
  useEffect(() => {
    setText(values.join('\n'));
  }, [values]);
  return (
    <div>
      <label className="block text-xs font-medium text-[#615d59] mb-1">{label}</label>
      <textarea
        value={text}
        onChange={(e) => {
          setText(e.target.value);
          onChange(e.target.value.split('\n').filter((s) => s.trim()));
        }}
        rows={3}
        className="w-full px-2.5 py-2 border border-black/10 rounded text-sm bg-white text-black/95 outline-none focus:border-[#0075de] transition-colors resize-none"
        placeholder="每行一个"
      />
    </div>
  );
}
