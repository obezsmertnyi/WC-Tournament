/**
 * Localized national-team names keyed by FIFA 3-letter code.
 *
 * The API returns team names in English ("Mexico", "Korea Republic", …). To
 * localize them on language switch we keep a per-code map of names here:
 *   - `uk` — the Ukrainian name (required when a code is listed).
 *   - `en` — an optional English override. When omitted we fall back to the
 *     name the API provided, so EN never goes blank even for codes we miss.
 *
 * Country names are intentionally kept here (not in the i18n JSON) because they
 * are data, keyed by a stable code, not free-form UI copy.
 *
 * Coverage: every team currently in the dataset plus the common WC nations, so
 * names resolve for the full 48-team field. Unknown codes fall back gracefully
 * to the API English name via `teamName()`.
 */

export interface TeamNameEntry {
  /** Optional English override; falls back to the API name when omitted. */
  en?: string
  /** Ukrainian name. */
  uk: string
}

export const TEAM_NAMES: Record<string, TeamNameEntry> = {
  // ── Hosts + current dataset ───────────────────────────────────────────────
  USA: { en: 'USA', uk: 'США' },
  CAN: { en: 'Canada', uk: 'Канада' },
  MEX: { en: 'Mexico', uk: 'Мексика' },
  ARG: { en: 'Argentina', uk: 'Аргентина' },
  BRA: { en: 'Brazil', uk: 'Бразилія' },
  FRA: { en: 'France', uk: 'Франція' },
  ENG: { en: 'England', uk: 'Англія' },
  ESP: { en: 'Spain', uk: 'Іспанія' },
  GER: { en: 'Germany', uk: 'Німеччина' },
  POR: { en: 'Portugal', uk: 'Португалія' },
  NED: { en: 'Netherlands', uk: 'Нідерланди' },
  BEL: { en: 'Belgium', uk: 'Бельгія' },
  CRO: { en: 'Croatia', uk: 'Хорватія' },
  URU: { en: 'Uruguay', uk: 'Уругвай' },
  COL: { en: 'Colombia', uk: 'Колумбія' },
  JPN: { en: 'Japan', uk: 'Японія' },
  AUS: { en: 'Australia', uk: 'Австралія' },
  SUI: { en: 'Switzerland', uk: 'Швейцарія' },
  MAR: { en: 'Morocco', uk: 'Марокко' },
  SEN: { en: 'Senegal', uk: 'Сенегал' },
  SRB: { en: 'Serbia', uk: 'Сербія' },
  GHA: { en: 'Ghana', uk: 'Гана' },
  CMR: { en: 'Cameroon', uk: 'Камерун' },
  ALG: { en: 'Algeria', uk: 'Алжир' },
  CIV: { en: "Côte d'Ivoire", uk: "Кот-д'Івуар" },
  BIH: { en: 'Bosnia and Herzegovina', uk: 'Боснія і Герцеговина' },
  COD: { en: 'Congo DR', uk: 'ДР Конго' },
  CPV: { en: 'Cabo Verde', uk: 'Кабо-Верде' },
  AUT: { en: 'Austria', uk: 'Австрія' },
  ITA: { en: 'Italy', uk: 'Італія' },
  QAT: { en: 'Qatar', uk: 'Катар' },
  IRN: { en: 'IR Iran', uk: 'Іран' },
  KSA: { en: 'Saudi Arabia', uk: 'Саудівська Аравія' },
  RSA: { en: 'South Africa', uk: 'Південна Африка' },
  KOR: { en: 'Korea Republic', uk: 'Південна Корея' },
  CZE: { en: 'Czechia', uk: 'Чехія' },
  PAR: { en: 'Paraguay', uk: 'Парагвай' },
  HAI: { en: 'Haiti', uk: 'Гаїті' },
  SCO: { en: 'Scotland', uk: 'Шотландія' },
  TUR: { en: 'Türkiye', uk: 'Туреччина' },
  CUW: { en: 'Curaçao', uk: 'Кюрасао' },
  ECU: { en: 'Ecuador', uk: 'Еквадор' },
  SWE: { en: 'Sweden', uk: 'Швеція' },
  TUN: { en: 'Tunisia', uk: 'Туніс' },
  EGY: { en: 'Egypt', uk: 'Єгипет' },
  NZL: { en: 'New Zealand', uk: 'Нова Зеландія' },
  IRQ: { en: 'Iraq', uk: 'Ірак' },
  NOR: { en: 'Norway', uk: 'Норвегія' },
  JOR: { en: 'Jordan', uk: 'Йорданія' },
  PAN: { en: 'Panama', uk: 'Панама' },
  UZB: { en: 'Uzbekistan', uk: 'Узбекистан' },

  // ── Home nations ─────────────────────────────────────────────────────────
  WAL: { en: 'Wales', uk: 'Уельс' },
  NIR: { en: 'Northern Ireland', uk: 'Північна Ірландія' },

  // ── Europe ───────────────────────────────────────────────────────────────
  POL: { en: 'Poland', uk: 'Польща' },
  UKR: { en: 'Ukraine', uk: 'Україна' },
  DEN: { en: 'Denmark', uk: 'Данія' },
  FIN: { en: 'Finland', uk: 'Фінляндія' },
  ISL: { en: 'Iceland', uk: 'Ісландія' },
  IRL: { en: 'Ireland', uk: 'Ірландія' },
  GRE: { en: 'Greece', uk: 'Греція' },
  RUS: { en: 'Russia', uk: 'Росія' },
  ROU: { en: 'Romania', uk: 'Румунія' },
  HUN: { en: 'Hungary', uk: 'Угорщина' },
  SVK: { en: 'Slovakia', uk: 'Словаччина' },
  SVN: { en: 'Slovenia', uk: 'Словенія' },
  BUL: { en: 'Bulgaria', uk: 'Болгарія' },
  ALB: { en: 'Albania', uk: 'Албанія' },
  MKD: { en: 'North Macedonia', uk: 'Північна Македонія' },
  MNE: { en: 'Montenegro', uk: 'Чорногорія' },
  GEO: { en: 'Georgia', uk: 'Грузія' },
  ARM: { en: 'Armenia', uk: 'Вірменія' },
  AZE: { en: 'Azerbaijan', uk: 'Азербайджан' },
  BLR: { en: 'Belarus', uk: 'Білорусь' },
  LUX: { en: 'Luxembourg', uk: 'Люксембург' },
  KOS: { en: 'Kosovo', uk: 'Косово' },

  // ── Americas ─────────────────────────────────────────────────────────────
  CHI: { en: 'Chile', uk: 'Чилі' },
  PER: { en: 'Peru', uk: 'Перу' },
  VEN: { en: 'Venezuela', uk: 'Венесуела' },
  BOL: { en: 'Bolivia', uk: 'Болівія' },
  CRC: { en: 'Costa Rica', uk: 'Коста-Рика' },
  HON: { en: 'Honduras', uk: 'Гондурас' },
  JAM: { en: 'Jamaica', uk: 'Ямайка' },
  SLV: { en: 'El Salvador', uk: 'Сальвадор' },
  GUA: { en: 'Guatemala', uk: 'Гватемала' },
  TRI: { en: 'Trinidad and Tobago', uk: 'Тринідад і Тобаго' },
  SUR: { en: 'Suriname', uk: 'Суринам' },

  // ── Africa ───────────────────────────────────────────────────────────────
  NGA: { en: 'Nigeria', uk: 'Нігерія' },
  MLI: { en: 'Mali', uk: 'Малі' },
  BFA: { en: 'Burkina Faso', uk: 'Буркіна-Фасо' },
  GUI: { en: 'Guinea', uk: 'Гвінея' },
  GAB: { en: 'Gabon', uk: 'Габон' },
  ZAM: { en: 'Zambia', uk: 'Замбія' },
  ANG: { en: 'Angola', uk: 'Ангола' },
  UGA: { en: 'Uganda', uk: 'Уганда' },
  KEN: { en: 'Kenya', uk: 'Кенія' },
  NAM: { en: 'Namibia', uk: 'Намібія' },
  ZIM: { en: 'Zimbabwe', uk: 'Зімбабве' },
  BEN: { en: 'Benin', uk: 'Бенін' },
  TOG: { en: 'Togo', uk: 'Того' },

  // ── Asia / Oceania ───────────────────────────────────────────────────────
  UAE: { en: 'United Arab Emirates', uk: 'ОАЕ' },
  OMA: { en: 'Oman', uk: 'Оман' },
  CHN: { en: 'China PR', uk: 'Китай' },
  PRK: { en: 'Korea DPR', uk: 'КНДР' },
  THA: { en: 'Thailand', uk: 'Таїланд' },
  VIE: { en: 'Vietnam', uk: 'Вʼєтнам' },
  IDN: { en: 'Indonesia', uk: 'Індонезія' },
  IND: { en: 'India', uk: 'Індія' },
  FIJ: { en: 'Fiji', uk: 'Фіджі' },
  SYR: { en: 'Syria', uk: 'Сирія' },
  LBN: { en: 'Lebanon', uk: 'Ліван' },
  PLE: { en: 'Palestine', uk: 'Палестина' },
}

/**
 * Resolve a localized team name.
 *   - lang === 'uk': the Ukrainian name when known, else the API name.
 *   - otherwise: the English override when set, else the API name.
 * Always returns a non-empty string when `apiName` is provided, so the UI never
 * shows a blank or crashes on an unmapped code.
 */
export function teamName(
  code: string | undefined | null,
  apiName: string | undefined | null,
  lang: string | undefined,
): string {
  const fallback = (apiName ?? '').trim()
  const key = (code ?? '').trim().toUpperCase()
  const entry = key ? TEAM_NAMES[key] : undefined

  if (lang === 'uk') {
    return entry?.uk || fallback || key
  }
  return entry?.en || fallback || key
}
