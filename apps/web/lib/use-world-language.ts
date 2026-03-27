"use client";

import { useEffect, useState } from "react";

import { storageKey, type Language } from "./world-ui";

export function useWorldLanguage(initialLanguage: Language = "zh-CN") {
  const [language, setLanguage] = useState<Language>(initialLanguage);

  useEffect(() => {
    const savedLanguage = window.localStorage.getItem(storageKey);
    if (savedLanguage === "zh-CN" || savedLanguage === "en-US") {
      setLanguage(savedLanguage);
      document.documentElement.lang = savedLanguage;
      return;
    }

    document.documentElement.lang = initialLanguage;
  }, [initialLanguage]);

  useEffect(() => {
    document.documentElement.lang = language;
    window.localStorage.setItem(storageKey, language);
  }, [language]);

  return {
    language,
    toggleLanguage: () => setLanguage((current) => (current === "zh-CN" ? "en-US" : "zh-CN")),
  };
}
