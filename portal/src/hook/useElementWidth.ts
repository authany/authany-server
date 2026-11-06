import { RefObject, useLayoutEffect, useState } from "react";

// Width of an element in px, kept current with a ResizeObserver. Lets a list
// lay itself out by its own width, not the viewport's (the sidebar takes part
// of the viewport).
export function useElementWidth(ref: RefObject<HTMLElement | null>): number {
  const [width, setWidth] = useState(0);
  useLayoutEffect(() => {
    const el = ref.current;
    if (el == null) {
      return undefined;
    }
    setWidth(el.getBoundingClientRect().width);
    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setWidth(entry.contentRect.width);
      }
    });
    observer.observe(el);
    return () => observer.disconnect();
  }, [ref]);
  return width;
}
