import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import en from './locales/en.json'
import uk from './locales/uk.json'

export const SUPPORTED_LANGUAGES = ['uk', 'en'] as const
export type Language = (typeof SUPPORTED_LANGUAGES)[number]

export const resources = {
  en: { translation: en },
  uk: { translation: uk },
} as const

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'uk',
    supportedLngs: SUPPORTED_LANGUAGES,
    // Only ever resolve to the base language (e.g. "en-US" -> "en").
    load: 'languageOnly',
    nonExplicitSupportedLngs: true,
    detection: {
      // Respect a saved choice first, then the browser preference.
      order: ['localStorage', 'navigator'],
      lookupLocalStorage: 'wc-lang',
      caches: ['localStorage'],
    },
    interpolation: {
      escapeValue: false,
    },
  })

/** Keep <html lang> in sync with the active language. */
function syncHtmlLang(lng: string) {
  document.documentElement.lang = lng
}
syncHtmlLang(i18n.resolvedLanguage ?? i18n.language)
i18n.on('languageChanged', syncHtmlLang)

export default i18n
