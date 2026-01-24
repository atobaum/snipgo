import { useState, useMemo, useCallback, useEffect, useRef } from "react";

export interface FilterableDropdownState<T> {
  isOpen: boolean;
  setIsOpen: (value: boolean) => void;
  filter: string;
  setFilter: (value: string) => void;
  highlightedIndex: number;
  setHighlightedIndex: (value: number) => void;
  filteredItems: T[];
  dropdownRef: React.RefObject<HTMLDivElement>;
  handleKeyDown: (e: React.KeyboardEvent, onSelect: (item: T) => void) => void;
  reset: () => void;
}

/**
 * Hook for filterable dropdown with keyboard navigation
 * @param items - All available items
 * @param filterFn - Function to filter items based on filter string
 * @param initialValue - Initial value to show when closed
 */
export function useFilterableDropdown<T>(
  items: T[],
  filterFn: (item: T, filter: string) => boolean,
  initialValue: string = ""
): FilterableDropdownState<T> {
  const [isOpen, setIsOpen] = useState(false);
  const [filter, setFilter] = useState("");
  const [highlightedIndex, setHighlightedIndex] = useState(-1);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Filter items based on current filter
  const filteredItems = useMemo(() => {
    if (!filter) return items;
    return items.filter((item) => filterFn(item, filter));
  }, [items, filter, filterFn]);

  // Scroll highlighted item into view
  useEffect(() => {
    if (highlightedIndex >= 0 && dropdownRef.current) {
      const el = dropdownRef.current.querySelector(
        `[data-dropdown-index="${highlightedIndex}"]`
      );
      el?.scrollIntoView({ block: "nearest" });
    }
  }, [highlightedIndex]);

  // Reset state when opening/closing
  const reset = useCallback(() => {
    setFilter(initialValue);
    setHighlightedIndex(-1);
  }, [initialValue]);

  // Handle keyboard navigation
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent, onSelect: (item: T) => void) => {
      if (e.key === "Escape" || e.key === "Tab") {
        setIsOpen(false);
        reset();
      } else if (e.key === "ArrowDown") {
        e.preventDefault();
        setHighlightedIndex((prev) =>
          prev < filteredItems.length - 1 ? prev + 1 : prev
        );
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setHighlightedIndex((prev) => (prev > 0 ? prev - 1 : prev));
      } else if (e.key === "Enter" && filteredItems.length > 0) {
        e.preventDefault();
        const selectedIndex = highlightedIndex >= 0 ? highlightedIndex : 0;
        onSelect(filteredItems[selectedIndex]);
        setIsOpen(false);
        reset();
      }
    },
    [filteredItems, highlightedIndex, reset]
  );

  return {
    isOpen,
    setIsOpen,
    filter,
    setFilter,
    highlightedIndex,
    setHighlightedIndex,
    filteredItems,
    dropdownRef,
    handleKeyDown,
    reset,
  };
}
