'use client';

import { useState } from "react";
import Link from "next/link";

export default function NavbarLogo() {
  const [clickCount, setClickCount] = useState(0);

  const spinSheep = () => {
    setClickCount(prev => prev + 1);
  };

  return (
    <Link
      href="/"
      className="flex items-center gap-3 group select-none transition-transform active:scale-95"
    >
      <span
        onClick={(e) => {
          e.preventDefault();
          spinSheep();
        }}
        className="text-3xl transition-transform duration-700 ease-in-out cursor-pointer"
        style={{ transform: `rotate(${clickCount * 360}deg)` }}
      >
        🐑
      </span>
      <span className="text-xl font-bold tracking-tight font-mono group-hover:text-yapi-accent transition-colors">yapi</span>
    </Link>
  );
}
