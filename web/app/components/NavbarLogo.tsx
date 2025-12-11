'use client';

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function NavbarLogo() {
  const [clickCount, setClickCount] = useState(0);
  const router = useRouter();

  const spinAndGoHome = () => {
    setClickCount(prev => prev + 1);
    setTimeout(() => {
      router.push("/");
    }, 700);
  };

  return (
    <button
      onClick={spinAndGoHome}
      className="flex items-center gap-3 group select-none transition-transform active:scale-95"
    >
      <span
        className="text-3xl transition-transform duration-700 ease-in-out cursor-pointer"
        style={{ transform: `rotate(${clickCount * 360}deg)` }}
      >
        🐑
      </span>
      <span className="text-xl font-bold tracking-tight font-mono group-hover:text-yapi-accent transition-colors">yapi</span>
    </button>
  );
}
