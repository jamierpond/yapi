'use client';

import { useState } from "react";

export default function NavbarLogo() {
  const [clickCount, setClickCount] = useState(0);

  const spinSheep = () => {
    setClickCount(prev => prev + 1);
  };

  return (
    <button
      onClick={spinSheep}
      className="flex items-center gap-3 group select-none transition-transform active:scale-95"
    >
      <span
        className="text-3xl transition-transform duration-700 ease-in-out"
        style={{ transform: `rotate(${clickCount * 360}deg)` }}
      >
        🐑
      </span>
      <span className="text-xl font-bold tracking-tight font-mono group-hover:text-yapi-accent transition-colors">yapi</span>
    </button>
  );
}
