import { useEffect } from "react";
import { useThemeStore } from "@tableflow/ui-library";
import useSearchParams from "../hooks/useSearchParams";
import useEmbedStore from "../stores/embed";
import parseCssOverrides from "../utils/cssInterpreter";
import { EmbedProps } from "./types";

export default function Embed({ children }: EmbedProps) {
  const {
    importerId,
    darkMode: darkModeString,
    primaryColor,
    metadata,
    template,
    isOpen,
    onComplete,
    customStyles,
    showImportLoadingStatus,
    skipHeaderRowSelection,
    cssOverrides,
    isModal,
    schemaless,
  } = useSearchParams();

  // Set importerId & metadata in embed store
  const setEmbedParams = useEmbedStore((state) => state.setEmbedParams);
  const strToBoolean = (str: string) => !!str && (str.toLowerCase() === "true" || str === "1");
  const strToOptionalBoolean = (str: string) => (str ? str.toLowerCase() === "true" || str === "1" : undefined);
  const validateJSON = (str: string) => {
    if (!str) {
      return "";
    }
    try {
      const obj = JSON.parse(str);
      return JSON.stringify(obj);
    } catch (e) {
      return "";
    }
  };

  useEffect(() => {
    setEmbedParams({
      importerId,
      metadata: validateJSON(metadata),
      template: validateJSON(template),
      isOpen: strToBoolean(isOpen),
      onComplete: strToBoolean(onComplete),
      showImportLoadingStatus: strToBoolean(showImportLoadingStatus),
      skipHeaderRowSelection: strToOptionalBoolean(skipHeaderRowSelection),
      isModal: strToOptionalBoolean(isModal),
      schemaless: strToOptionalBoolean(schemaless),
    });
  }, [importerId, metadata]);

  // Set Light/Dark mode
  const darkMode = strToBoolean(darkModeString);
  const setTheme = useThemeStore((state) => state.setTheme);

  useEffect(() => {
    setTheme(darkMode ? "dark" : "light");
  }, [darkMode]);

  // Apply primary color
  useEffect(() => {
    if (primaryColor) {
      const root = document.documentElement;
      root.style.setProperty("--color-primary", primaryColor);
    }
  }, [primaryColor]);

  // Apply custom CSS properties
  useEffect(() => {
    try {
      if (customStyles && customStyles !== "undefined") {
        const parsedStyles = JSON.parse(customStyles);

        if (customStyles && parsedStyles) {
          Object.keys(parsedStyles).forEach((key) => {
            const root = document.documentElement;
            const value = parsedStyles?.[key as any];
            root.style.setProperty("--" + key, value);
          });
        }
      }
    } catch (e) {
      console.error('The "customStyles" prop is not a valid JSON string. Please check the documentation for more details.');
    }
  }, [customStyles]);

  // Apply CSS overrides
  useEffect(() => {
    const parsedCss = parseCssOverrides(cssOverrides);

    if (parsedCss) {
      const style = document.createElement("style");
      style.textContent = decodeURIComponent(parsedCss);
      document.head.append(style);
    }
  }, [cssOverrides]);

  return <>{children}</>;
}
