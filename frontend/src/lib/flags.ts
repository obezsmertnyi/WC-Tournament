/**
 * FIFA three-letter country codes → ISO 3166-1 alpha-2 codes used by the
 * `flag-icons` CSS package (which expects lowercase two-letter codes, plus a
 * few special subdivision codes like `gb-eng` for England).
 *
 * The map covers every team in the current dataset plus a broad set of common
 * nations so flags resolve for the full 48-team field. When a code is missing
 * here the <Flag> component falls back to the API-provided image, then to a
 * neutral monogram chip.
 */
export const FIFA_TO_ISO: Record<string, string> = {
  // Hosts + dataset teams
  MEX: 'mx',
  RSA: 'za',
  KOR: 'kr',
  CZE: 'cz',
  USA: 'us',
  CAN: 'ca',
  ARG: 'ar',
  BRA: 'br',
  FRA: 'fr',
  ENG: 'gb-eng',
  ESP: 'es',
  GER: 'de',
  POR: 'pt',
  NED: 'nl',
  BEL: 'be',
  CRO: 'hr',
  URU: 'uy',
  COL: 'co',
  JPN: 'jp',
  AUS: 'au',
  SUI: 'ch',
  MAR: 'ma',
  SEN: 'sn',
  SRB: 'rs',
  GHA: 'gh',
  CMR: 'cm',
  ALG: 'dz',
  CIV: 'ci',
  BIH: 'ba',
  COD: 'cd',
  CPV: 'cv',
  AUT: 'at',
  ITA: 'it',
  // Home nations
  SCO: 'gb-sct',
  WAL: 'gb-wls',
  NIR: 'gb-nir',
  // Europe
  POL: 'pl',
  UKR: 'ua',
  DEN: 'dk',
  SWE: 'se',
  NOR: 'no',
  FIN: 'fi',
  ISL: 'is',
  IRL: 'ie',
  GRE: 'gr',
  TUR: 'tr',
  RUS: 'ru',
  ROU: 'ro',
  HUN: 'hu',
  SVK: 'sk',
  SVN: 'si',
  BUL: 'bg',
  ALB: 'al',
  MKD: 'mk',
  MNE: 'me',
  GEO: 'ge',
  ARM: 'am',
  AZE: 'az',
  BLR: 'by',
  LUX: 'lu',
  KOS: 'xk',
  // Americas
  CHI: 'cl',
  PAR: 'py',
  PER: 'pe',
  ECU: 'ec',
  VEN: 've',
  BOL: 'bo',
  CRC: 'cr',
  PAN: 'pa',
  HON: 'hn',
  JAM: 'jm',
  HAI: 'ht',
  SLV: 'sv',
  GUA: 'gt',
  TRI: 'tt',
  CUW: 'cw',
  SUR: 'sr',
  // Africa
  NGA: 'ng',
  EGY: 'eg',
  TUN: 'tn',
  MLI: 'ml',
  BFA: 'bf',
  GUI: 'gn',
  GAB: 'ga',
  ZAM: 'zm',
  ANG: 'ao',
  MOZ: 'mz',
  UGA: 'ug',
  KEN: 'ke',
  TAN: 'tz',
  GAM: 'gm',
  MTN: 'mr',
  NAM: 'na',
  ZIM: 'zw',
  BEN: 'bj',
  TOG: 'tg',
  // Asia / Oceania
  KSA: 'sa',
  IRN: 'ir',
  IRQ: 'iq',
  QAT: 'qa',
  UAE: 'ae',
  JOR: 'jo',
  OMA: 'om',
  CHN: 'cn',
  PRK: 'kp',
  THA: 'th',
  VIE: 'vn',
  IDN: 'id',
  MAS: 'my',
  UZB: 'uz',
  IND: 'in',
  NZL: 'nz',
  FIJ: 'fj',
  PNG: 'pg',
  SYR: 'sy',
  LBN: 'lb',
  PLE: 'ps',
}

/**
 * Resolve a FIFA-3 (or already-ISO-2) code to a `flag-icons` ISO code.
 * Returns `undefined` when no mapping is known so callers can fall back.
 */
export function fifaToIso(code: string | undefined | null): string | undefined {
  if (!code) return undefined
  const raw = code.trim()
  if (!raw) return undefined
  const upper = raw.toUpperCase()
  if (FIFA_TO_ISO[upper]) return FIFA_TO_ISO[upper]
  // Already a 2-letter ISO code (e.g. backend sends "mx").
  if (raw.length === 2) return raw.toLowerCase()
  return undefined
}
