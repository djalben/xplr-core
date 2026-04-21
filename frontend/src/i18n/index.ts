import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

import ru from './locales/ru.json';
import en from './locales/en.json';
import es from './locales/es.json';
import pt from './locales/pt.json';
import tr from './locales/tr.json';
import zh from './locales/zh.json';

const LANG_KEY = 'xplr_language';

i18n.use(initReactI18next).init({
  resources: {
    ru: { translation: ru },
    en: { translation: en },
    es: { translation: es },
    pt: { translation: pt },
    tr: { translation: tr },
    zh: { translation: zh },
  },
  lng: localStorage.getItem(LANG_KEY) || 'ru',
  fallbackLng: 'ru',
  interpolation: { escapeValue: false },
});

export default i18n;
export { LANG_KEY };
