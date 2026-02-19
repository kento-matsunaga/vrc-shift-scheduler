import { useState, useRef, useCallback, useEffect } from 'react';
import { createPortal } from 'react-dom';

interface NoteTooltipProps {
  note: string;
  className?: string;
}

export default function NoteTooltip({ note, className = '' }: NoteTooltipProps) {
  const [visible, setVisible] = useState(false);
  const [position, setPosition] = useState<{ top: number; left: number; showAbove: boolean } | null>(null);
  const triggerRef = useRef<HTMLParagraphElement>(null);

  const updatePosition = useCallback(() => {
    if (!triggerRef.current) return;
    const rect = triggerRef.current.getBoundingClientRect();
    const cardHeight = 80;
    const showAbove = rect.top > cardHeight;

    // Use fixed positioning with viewport-relative coordinates
    const top = showAbove ? rect.top - 8 : rect.bottom + 8;

    // Center horizontally, clamp to viewport
    let left = rect.left + rect.width / 2;
    const margin = 16;
    const halfCard = 160; // half of max-w-xs (320px)
    if (left < halfCard + margin) left = halfCard + margin;
    if (left > window.innerWidth - halfCard - margin) left = window.innerWidth - halfCard - margin;

    setPosition({ top, left, showAbove });
  }, []);

  const handleMouseEnter = useCallback(() => {
    updatePosition();
    setVisible(true);
  }, [updatePosition]);

  const handleMouseLeave = useCallback(() => {
    setVisible(false);
  }, []);

  // Mobile: toggle on tap
  const handleClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    if (visible) {
      setVisible(false);
    } else {
      updatePosition();
      setVisible(true);
    }
  }, [visible, updatePosition]);

  // Close on scroll, resize, or outside tap
  useEffect(() => {
    if (!visible) return;

    const handleClose = () => setVisible(false);
    window.addEventListener('scroll', handleClose, true);
    window.addEventListener('resize', handleClose);

    return () => {
      window.removeEventListener('scroll', handleClose, true);
      window.removeEventListener('resize', handleClose);
    };
  }, [visible]);

  return (
    <>
      <p
        ref={triggerRef}
        className={`line-clamp-2 break-words cursor-pointer ${className}`}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        onClick={handleClick}
      >
        {note}
      </p>
      {visible && position && createPortal(
        <div
          className="fixed z-50 max-w-xs bg-white rounded-lg shadow-lg border border-gray-200 px-3 py-2 text-xs text-gray-700 leading-relaxed break-words pointer-events-none"
          style={{
            top: position.top,
            left: position.left,
            transform: position.showAbove
              ? 'translate(-50%, -100%)'
              : 'translate(-50%, 0)',
          }}
        >
          {note}
        </div>,
        document.body
      )}
    </>
  );
}
