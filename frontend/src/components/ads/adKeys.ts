// All ad-network codes for chit-nya.pages.dev live here. Codes are per-website:
// they come from the ChitNya website in the ad network dashboard (GET CODE).
// To turn a slot off, set its key to '' (or the ENABLE_* flag to false).

// Banner (iframe) units — served from this host.
export const AD_HOST = 'https://www.highrevenueformat.com'
export const KEY_320x50  = '12bf7f34bf83c6c9826ed7519f6c8421' // phone: sticky bar at the bottom
export const KEY_300x250 = 'e0829f8d97333ecb257fa73dc99212d3' // phone + desktop: below the player
export const KEY_728x90  = 'e526e41c2a537a2ffb0826042427ac78' // desktop only
export const KEY_160x600 = '0336c708e08e3ab9b07089c848fa6f1d' // desktop only: left/right of the page

// Native banner (a script + a container whose id contains the key).
// Disabled: this format runs an unsandboxed <script> in the page (not an
// iframe), so it can — and on this account did — render a full-page fake
// "notification" overlay that hijacks clicks anywhere, same as the pop-under.
export const ENABLE_NATIVE_BANNER = false
export const NATIVE_KEY = '59da13e29228d10fd398f280685fd39d'
export const NATIVE_SRC = `https://pl31448491.profitableratecpmnetwork.com/${NATIVE_KEY}/invoke.js`

// Pop-under and Social Bar: one script each, loaded once for the whole site.
export const ENABLE_POPUNDER = true
export const POPUNDER_SRC = 'https://pl31448490.profitableratecpmnetwork.com/a5/30/af/a530af95b1e952fc64775ac92878418f.js'
export const ENABLE_SOCIAL_BAR = true
export const SOCIAL_BAR_SRC = 'https://pl31448493.profitableratecpmnetwork.com/bc/55/b4/bc55b40db604f5298b33044d98fd0a82.js'

// Smartlink: shown as a visibly labelled "Sponsored" button (never a hidden redirect).
export const SMARTLINK_URL = 'https://www.profitableratecpmnetwork.com/sjz4qyf3as?key=eec19deac7350fbe61f0c714110aaa6d'

// Side ads on wide desktops (>= 1780px): both sides use the 160x600 unit.
export const SIDE_LEFT_KEY = KEY_160x600
export const SIDE_RIGHT_KEY = KEY_160x600
export const SIDE_WIDTH = 160
export const SIDE_HEIGHT = 600
