'use client';

export default function LandingStyles() {
  return (
    <style jsx global>{`
      @keyframes shine {
        to {
          background-position: 200% center;
        }
      }
      .animate-shine {
        animation: shine 4s linear infinite;
      }
      .animate-pulse-slow {
        animation: pulse 4s cubic-bezier(0.4, 0, 0.6, 1) infinite;
      }
      .perspective-1000 {
        perspective: 1000px;
      }
    `}</style>
  );
}
