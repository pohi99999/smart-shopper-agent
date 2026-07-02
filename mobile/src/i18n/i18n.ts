import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import * as Localization from 'expo-localization';

import hu from '../locales/hu.json';
import en from '../locales/en.json';

const resources = {
  hu: { translation: hu },
  en: { translation: en },
};

const getLocale = (): string => {
  const locales = Localization.getLocales();
  if (locales && locales.length > 0 && locales[0].languageCode) {
    return locales[0].languageCode;
  }
  return 'hu';
};

i18n
  .use(initReactI18next)
  .init({
    resources,
    lng: getLocale(),
    fallbackLng: 'hu',
    compatibilityJSON: 'v3',
    interpolation: {
      escapeValue: false,
    },
  });

export default i18n;
