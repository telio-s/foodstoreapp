import type { ButtonHTMLAttributes } from "react";

type ButtonVariant = "pill" | "circle";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
};

const variantClassName: Record<ButtonVariant, string> = {
  pill: "w-full rounded-full py-4 font-medium",
  circle: "flex h-9 w-9 items-center justify-center rounded-full",
};

export function Button({
  variant = "pill",
  type = "button",
  className = "",
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      className={`bg-blue-400 text-white transition-colors hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-50 ${variantClassName[variant]} ${className}`}
      {...props}
    />
  );
}
