import { useRef, useCallback } from "react";
import { useClickOutside } from "../hooks/useClickOutside";
import { useFilterableDropdown } from "../hooks/useFilterableDropdown";
import { SUPPORTED_LANGUAGES } from "../constants/editor";

interface LanguageSelectorProps {
  language: string;
  onLanguageChange: (language: string) => void;
}

export function LanguageSelector({
  language,
  onLanguageChange,
}: LanguageSelectorProps) {
  const containerRef = useRef<HTMLDivElement>(null);

  const {
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
  } = useFilterableDropdown(
    SUPPORTED_LANGUAGES,
    (lang, f) => lang.toLowerCase().includes(f.toLowerCase()),
    language
  );

  useClickOutside(
    containerRef,
    useCallback(() => {
      setIsOpen(false);
      reset();
    }, [setIsOpen, reset]),
    isOpen
  );

  const handleSelect = (lang: string) => {
    onLanguageChange(lang);
    setIsOpen(false);
    reset();
  };

  return (
    <div ref={containerRef} className="relative">
      <label className="mr-2">Language:</label>
      <input
        type="text"
        autoCorrect="off"
        autoCapitalize="off"
        spellCheck={false}
        value={isOpen ? filter : language}
        onChange={(e) => {
          setFilter(e.target.value);
          setHighlightedIndex(-1);
          if (!isOpen) setIsOpen(true);
        }}
        onFocus={() => {
          setFilter(language);
          setIsOpen(true);
          setHighlightedIndex(-1);
        }}
        onKeyDown={(e) => handleKeyDown(e, handleSelect)}
        placeholder="Select..."
        className="px-2 py-1 border border-gray-300 rounded w-32 text-xs"
      />
      {isOpen && (
        <div
          ref={dropdownRef}
          className="absolute left-16 top-full mt-1 bg-white border border-gray-200 rounded shadow-lg z-20 max-h-48 overflow-y-auto min-w-[140px]"
        >
          {filteredItems.length > 0 ? (
            filteredItems.map((lang, index) => (
              <button
                key={lang}
                data-dropdown-index={index}
                onClick={() => handleSelect(lang)}
                className={`w-full px-3 py-1.5 text-left text-xs hover:bg-blue-50 ${
                  index === highlightedIndex
                    ? "bg-blue-100 text-blue-800"
                    : language === lang
                      ? "bg-blue-50 text-blue-800"
                      : "text-gray-700"
                }`}
              >
                {lang}
              </button>
            ))
          ) : (
            <div className="px-3 py-2 text-xs text-gray-500">
              Press Enter to use &quot;{filter}&quot;
            </div>
          )}
          {filter &&
            !filteredItems.includes(filter.toLowerCase()) &&
            filteredItems.length > 0 && (
              <button
                onClick={() => handleSelect(filter)}
                className="w-full px-3 py-1.5 text-left text-xs hover:bg-blue-50 text-blue-600 border-t border-gray-100"
              >
                Use &quot;{filter}&quot;
              </button>
            )}
        </div>
      )}
    </div>
  );
}
