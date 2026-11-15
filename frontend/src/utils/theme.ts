import "ace-builds";
import { themesByName } from "ace-builds/src-noconflict/ext-themelist";

export const getTheme = (): UserTheme => "light";

export const setTheme = (_theme: UserTheme) => {
  document.documentElement.classList.remove(_theme || "dark", "dark");
  document.documentElement.classList.add("light");
  document.documentElement.style.colorScheme = "light";
};

export const toggleTheme = (): void => setTheme("light");
export const getMediaPreference = (): UserTheme => "light";

export const getEditorTheme = (themeName: string) => {
  if (!themeName.startsWith("ace/theme/")) {
    themeName = `ace/theme/${themeName}`;
  }
  const themeKey = themeName.replace("ace/theme/", "");
  if (themesByName[themeKey] !== undefined && !themesByName[themeKey].isDark) {
    return themeName;
  } else {
    return "ace/theme/chrome";
  }
};
