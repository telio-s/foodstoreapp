import type { InputHTMLAttributes } from "react";

type TextInputProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
};

export function TextInput({
  label,
  id,
  className = "",
  ...props
}: TextInputProps) {
  const input = (
    <input
      id={id}
      className={`appearance-none rounded-full border-none bg-white px-4 py-3 text-black shadow-sm outline-none placeholder:text-zinc-400 ${className}`}
      {...props}
    />
  );

  if (!label) return input;

  return (
    <div className="flex flex-col gap-2">
      <label htmlFor={id} className="font-medium text-zinc-900">
        {label}
      </label>
      {input}
    </div>
  );
}
