import { ref } from 'vue'

export type Lang = 'my' | 'en'

const KEY = 'lang'

const messages: Record<Lang, Record<string, string>> = {
  en: {
    'common.back': 'Back',
    'common.tryAgain': 'Try again',
    'common.retry': 'Retry',
    'common.unknownError': 'Unknown error',
    'home.searchPlaceholder': 'Search…',
    'home.searchButton': 'Search',
    'home.all': 'All',
    'home.actresses': '🎭 Actresses',
    'home.recent': 'Recently watched',
    'home.clearHistory': 'Clear',
    'home.noResults': 'No results for "{q}"',
    'home.pageInfo': 'Page {page} of {total} ({count} videos)',
    'home.failedLoad': 'Failed to load videos',
    'card.new': 'NEW',
    'actors.title': 'Actresses',
    'actors.searchPlaceholder': 'Search actress…',
    'actors.noMatch': 'No actress found for "{q}"',
    'actors.noVideos': 'No videos for this actress',
    'actors.videoCount': '{n} videos',
    'pager.prev': '‹ Prev',
    'pager.next': 'Next ›',
    'pager.page': 'Page {p} / {t}',
    'watch.fetching': 'Fetching video…',
    'watch.sessionExpired': 'Video session expired or link opened in a new tab.',
    'watch.goHome': '← Go to Home',
    'watch.failedFetch': 'Failed to fetch video',
    'watch.category': 'Category:',
    'watch.actress': 'Actress:',
    'watch.tags': 'Tags:',
    'watch.serverFailed': 'Server {a} failed — trying Server {b}…',
    'watch.servers': 'Servers',
    'watch.linkCopied': 'Link Copied!',
    'watch.share': 'Share',
    'watch.next': 'Next:',
    'watch.playNow': 'Play now',
    'watch.related': 'Related videos',
    'share.loading': 'Loading video…',
    'share.notFound': 'Link not found',
    'share.goHome': 'Go to home',
    'player.auto': 'Auto',
    'player.playbackError': 'Playback error: {d}',
    'player.noHls': 'HLS is not supported in this browser.',
    'ad.label': 'Advertisement',
    'theme.toLight': 'Light mode',
    'theme.toDark': 'Dark mode',
    'lang.switch': 'Language',
  },
  my: {
    'common.back': 'နောက်သို့',
    'common.tryAgain': 'ပြန်စမ်းကြည့်ပါ',
    'common.retry': 'ပြန်စမ်းမည်',
    'common.unknownError': 'မသိသော အမှား',
    'home.searchPlaceholder': 'ရှာဖွေရန်…',
    'home.searchButton': 'ရှာ',
    'home.all': 'အားလုံး',
    'home.actresses': '🎭 မင်းသမီးများ',
    'home.recent': 'မကြာသေးမီ ကြည့်ခဲ့သော',
    'home.clearHistory': 'ဖျက်မည်',
    'home.noResults': '"{q}" အတွက် ရလဒ်မတွေ့ပါ',
    'home.pageInfo': 'စာမျက်နှာ {page} / {total} (ဗီဒီယို {count} ခု)',
    'home.failedLoad': 'ဗီဒီယိုများ မရယူနိုင်ပါ',
    'card.new': 'အသစ်',
    'actors.title': 'မင်းသမီးများ',
    'actors.searchPlaceholder': 'မင်းသမီး ရှာရန်…',
    'actors.noMatch': '"{q}" အတွက် မင်းသမီး မတွေ့ပါ',
    'actors.noVideos': 'ဒီမင်းသမီးအတွက် ဗီဒီယို မတွေ့ပါ',
    'actors.videoCount': 'ဗီဒီယို {n} ခု',
    'pager.prev': '‹ ရှေ့',
    'pager.next': 'နောက် ›',
    'pager.page': 'စာမျက်နှာ {p} / {t}',
    'watch.fetching': 'ဗီဒီယို ရယူနေသည်…',
    'watch.sessionExpired': 'ဗီဒီယို session ကုန်သွားပြီ၊ ဒါမှမဟုတ် link ကို tab အသစ်မှာ ဖွင့်ထားပါတယ်။',
    'watch.goHome': '← ပင်မစာမျက်နှာသို့',
    'watch.failedFetch': 'ဗီဒီယို မရယူနိုင်ပါ',
    'watch.category': 'အမျိုးအစား:',
    'watch.actress': 'မင်းသမီး:',
    'watch.tags': 'Tags:',
    'watch.serverFailed': 'Server {a} မအောင်မြင်ပါ — Server {b} ကို စမ်းနေသည်…',
    'watch.servers': 'Server များ',
    'watch.linkCopied': 'Link ကူးပြီးပြီ!',
    'watch.share': 'မျှဝေမည်',
    'watch.next': 'နောက်တစ်ခု:',
    'watch.playNow': 'ယခုကြည့်မည်',
    'watch.related': 'ဆင်တူသော ဗီဒီယိုများ',
    'share.loading': 'ဗီဒီယို ဖွင့်နေသည်…',
    'share.notFound': 'Link မတွေ့ပါ',
    'share.goHome': 'ပင်မစာမျက်နှာသို့',
    'player.auto': 'အလိုအလျောက်',
    'player.playbackError': 'ဖွင့်လို့မရပါ: {d}',
    'player.noHls': 'ဒီ browser မှာ HLS ကို မထောက်ပံ့ပါ။',
    'ad.label': 'ကြော်ငြာ',
    'theme.toLight': 'အလင်းမုဒ်',
    'theme.toDark': 'အမှောင်မုဒ်',
    'lang.switch': 'ဘာသာစကား',
  },
}

function stored(): Lang {
  try {
    return localStorage.getItem(KEY) === 'en' ? 'en' : 'my'
  } catch {
    return 'my'
  }
}

export const lang = ref<Lang>(stored())

function apply(l: Lang) {
  document.documentElement.lang = l === 'my' ? 'my' : 'en'
}

export function initLang() {
  apply(lang.value)
}

export function toggleLang() {
  lang.value = lang.value === 'my' ? 'en' : 'my'
  apply(lang.value)
  try {
    localStorage.setItem(KEY, lang.value)
  } catch { /* private mode — choice just isn't remembered */ }
}

// Reads lang.value, so anything calling t() inside a template re-renders on switch.
export function t(key: string, params?: Record<string, string | number>): string {
  let s = messages[lang.value][key] ?? messages.en[key] ?? key
  if (params) for (const [k, v] of Object.entries(params)) s = s.split(`{${k}}`).join(String(v))
  return s
}
