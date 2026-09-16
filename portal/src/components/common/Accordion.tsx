import cn from "classnames";
import React, { useCallback, useState } from "react";
import { Text } from "@radix-ui/themes";
import { ChevronDownIcon, ChevronUpIcon } from "@radix-ui/react-icons";
import styles from "./Accordion.module.css";

export function Accordion({
  className,
  text,
  defaultExpanded = false,
  children,
}: {
  className?: string;
  text: React.ReactNode;
  defaultExpanded?: boolean;
  children?: React.ReactNode;
}): React.ReactElement {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);

  const toggle = useCallback(() => {
    setIsExpanded((prev) => !prev);
  }, []);

  return (
    <div className={cn(className, styles.accordionRoot)}>
      <button
        className={styles.accordionToggle}
        type="button"
        onClick={toggle}
        aria-expanded={isExpanded}
      >
        <div className={styles.accordionToggleText}>
          <Text size="2">{text}</Text>
          {isExpanded ? (
            <ChevronUpIcon className={styles.accordionToggleIcon} />
          ) : (
            <ChevronDownIcon className={styles.accordionToggleIcon} />
          )}
        </div>
      </button>
      <div
        className={cn(
          styles.accordionContent,
          isExpanded ? null : styles["accordionContent--hide"]
        )}
      >
        {children}
      </div>
    </div>
  );
}
