import { structuredCloneSafe } from "./format.js";

export function emptyTariffDraft() {
  return {
    defaultCurrency: "usd",
    key: "",
    nameZh: "",
    nameEn: "",
    descriptionZh: "",
    descriptionEn: "",
    premiumNameZh: "",
    premiumNameEn: "",
    squadUuids: [],
    premiumSquadUuids: [],
    billing_model: "period",
    enabled: true,
    monthly_gb: 500,
    premium_monthly_gb: "",
    conversion_rate_rub_per_gb: "",
    periodRows: [
      { months: 1, rub: 3, referral_inviter: 3, referral_referee: 1 },
      { months: 3, rub: 9, referral_inviter: 7, referral_referee: 3 },
      { months: 6, rub: 18, referral_inviter: 15, referral_referee: 7 },
      { months: 12, rub: 36, referral_inviter: 30, referral_referee: 15 },
    ],
    topupRows: [],
    premiumTopupRows: [],
    trafficRows: [
      { gb: 10, price: 3 },
      { gb: 50, price: 10 },
    ],
  };
}

export function cloneCatalog(catalog) {
  return structuredCloneSafe({
    default_tariff: catalog?.default_tariff || "",
    default_currency: normalizeCurrencyKey(catalog?.default_currency || "usd"),
    topup_packages_default: catalog?.topup_packages_default || { usd: [], stars: [] },
    tariffs: catalog?.tariffs || [],
  });
}

export function normalizeCurrencyKey(value, fallback = "usd") {
  const text = String(value || "")
    .trim()
    .toLowerCase();
  if (!text) return fallback;
  if (text === "rur") return "rub";
  if (["xtr", "star", "stars"].includes(text)) return "stars";
  return text.replace(/[^a-z0-9_-]/g, "") || fallback;
}

export function rowsFromPackages(packageSet, currency, valueKey) {
  return (packageSet?.[currency] || []).map((pkg) => ({
    [valueKey]: pkg[valueKey],
    price: pkg.price,
    prices: pkg.prices ? structuredCloneSafe(pkg.prices) : undefined,
    min_price: pkg.min_price ?? "",
  }));
}

function packageValueSignature(value) {
  const num = Number(value);
  return Number.isFinite(num) ? String(num) : String(value || "");
}

export function packageRowsFromPackageSet(packageSet, currency, valueKey) {
  return (packageSet?.[currency] || []).map((pkg) => ({
    [valueKey]: pkg[valueKey],
    price: pkg.price,
    prices: pkg.prices ? structuredCloneSafe(pkg.prices) : undefined,
    min_price: pkg.min_price ?? "",
  }));
}

export function draftFromTariff(tariff, defaultCurrency = "usd") {
  const currency = normalizeCurrencyKey(defaultCurrency);
  const defaultPrices = tariff.prices?.[currency] || {};
  // enabled_periods comes first so its order (the configured purchase order)
  // is preserved; any extra price-only months are appended afterwards.
  const months = new Set([
    ...(tariff.enabled_periods || []),
    ...Object.keys(defaultPrices).map(Number),
    ...(currency === "rub" ? Object.keys(tariff.prices_rub || {}).map(Number) : []),
  ]);
  const periodRows = [...months]
    .filter((month) => Number.isFinite(month) && month > 0)
    .map((month) => ({
      months: month,
      rub:
        (currency === "rub" ? tariff.prices_rub?.[String(month)] : undefined) ??
        defaultPrices?.[String(month)] ??
        "",
      referral_inviter: tariff.referral_bonus_days_inviter?.[String(month)] ?? "",
      referral_referee: tariff.referral_bonus_days_referee?.[String(month)] ?? "",
    }));

  return {
    ...emptyTariffDraft(),
    defaultCurrency: currency,
    key: tariff.key || "",
    nameZh: tariff.names?.zh || "",
    nameEn: tariff.names?.en || "",
    descriptionZh: tariff.descriptions?.zh || "",
    descriptionEn: tariff.descriptions?.en || "",
    premiumNameZh: tariff.premium_names?.zh || "",
    premiumNameEn: tariff.premium_names?.en || "",
    squadUuids: tariff.squad_uuids || [],
    premiumSquadUuids: tariff.premium_squad_uuids || [],
    billing_model: tariff.billing_model || "period",
    enabled: tariff.enabled !== false,
    monthly_gb: tariff.monthly_gb ?? "",
    premium_monthly_gb: tariff.premium_monthly_gb ?? "",
    conversion_rate_rub_per_gb: tariff.conversion_rate_rub_per_gb ?? "",
    periodRows: periodRows.length ? periodRows : emptyTariffDraft().periodRows,
    topupRows: packageRowsFromPackageSet(tariff.topup_packages, currency, "gb"),
    premiumTopupRows: packageRowsFromPackageSet(tariff.premium_topup_packages, currency, "gb"),
    trafficRows: packageRowsFromPackageSet(tariff.traffic_packages, currency, "gb"),
  };
}

export function parseNumber(value, fallback = null) {
  if (value === "" || value === null || value === undefined) return fallback;
  const num = Number(value);
  return Number.isFinite(num) ? num : fallback;
}

export function parseIntNumber(value, fallback = null) {
  const num = parseNumber(value, fallback);
  return num === null ? fallback : Math.trunc(num);
}

export function compactMap(obj) {
  return Object.fromEntries(
    Object.entries(obj).filter(([, value]) => value !== "" && value !== null && value !== undefined)
  );
}

export function packagesFromRows(rows, valueKey) {
  const result = [];
  for (const row of rows || []) {
    const pkg = {
      [valueKey]: parseNumber(row[valueKey]),
      price: parseNumber(row.price),
    };
    if (row.prices && typeof row.prices === "object") {
      pkg.prices = structuredCloneSafe(row.prices);
    }
    const minPrice = parseNumber(row.min_price);
    if (minPrice !== null) {
      pkg.min_price = minPrice;
    }
    if (pkg[valueKey] > 0 && pkg.price !== null && pkg.price >= 0) {
      result.push(pkg);
    }
  }
  return result;
}

export function packagesFromPackageRows(rows, valueKey, priceKey) {
  return (rows || [])
    .map((row) => {
      const pkg = {
        [valueKey]: parseNumber(row[valueKey]),
        price: parseNumber(row[priceKey]),
      };
      return pkg;
    })
    .filter((row) => row[valueKey] > 0 && row.price !== null && row.price >= 0);
}

export function packageSetFromRows(rows, valueKey, defaultCurrency = "usd") {
  const currency = normalizeCurrencyKey(defaultCurrency);
  const defaultCurrencyPackages = packagesFromPackageRows(rows, valueKey, "price");
  if (!defaultCurrencyPackages.length) return null;
  return { [currency]: defaultCurrencyPackages };
}

export function normalizeUuidList(value) {
  if (Array.isArray(value)) return value.map((item) => String(item).trim()).filter(Boolean);
  return String(value || "")
    .split(/[\n,]+/)
    .map((item) => item.trim())
    .filter(Boolean);
}

export function tariffFromDraft(draft, fallbackCurrency = "usd") {
  const defaultCurrency = normalizeCurrencyKey(draft.defaultCurrency || fallbackCurrency);
  const key = draft.key.trim();
  const names = compactMap({ zh: draft.nameZh.trim(), en: draft.nameEn.trim() });
  const descriptions = compactMap({
    zh: draft.descriptionZh.trim(),
    en: draft.descriptionEn.trim(),
  });
  const premiumNames = compactMap({
    zh: draft.premiumNameZh.trim(),
    en: draft.premiumNameEn.trim(),
  });
  const tariff = {
    key,
    names,
    descriptions,
    premium_names: premiumNames,
    squad_uuids: normalizeUuidList(draft.squadUuids),
    premium_squad_uuids: normalizeUuidList(draft.premiumSquadUuids),
    billing_model: draft.billing_model,
    enabled: Boolean(draft.enabled),
  };

  const hwidLimit = parseIntNumber(draft.hwid_device_limit);
  if (hwidLimit !== null) tariff.hwid_device_limit = hwidLimit;
  const hwidPackages = packageSetFromRows(draft.hwidRows, "count", defaultCurrency);
  if (hwidPackages) tariff.hwid_device_packages = hwidPackages;
  const premiumMonthlyGb = parseNumber(draft.premium_monthly_gb);
  if (premiumMonthlyGb !== null) tariff.premium_monthly_gb = premiumMonthlyGb;
  const premiumTopupPackages = packageSetFromRows(draft.premiumTopupRows, "gb", defaultCurrency);
  if (premiumTopupPackages) tariff.premium_topup_packages = premiumTopupPackages;

  if (tariff.billing_model === "period") {
    const seenMonths = new Set();
    const rows = (draft.periodRows || [])
      .map((row) => ({
        months: parseIntNumber(row.months),
        rub: parseNumber(row.rub, 0),
        referral_inviter: parseIntNumber(row.referral_inviter),
        referral_referee: parseIntNumber(row.referral_referee),
      }))
      .filter((row) => row.months > 0)
      .filter((row) => {
        if (seenMonths.has(row.months)) return false;
        seenMonths.add(row.months);
        return true;
      });
    tariff.monthly_gb = parseNumber(draft.monthly_gb, 0);
    tariff.enabled_periods = rows.map((row) => row.months);
    const defaultPrices = Object.fromEntries(rows.map((row) => [String(row.months), row.rub || 0]));
    if (defaultCurrency === "rub") {
      tariff.prices_rub = defaultPrices;
    } else {
      tariff.prices = { [defaultCurrency]: defaultPrices };
    }
    tariff.referral_bonus_days_inviter = Object.fromEntries(
      rows
        .filter((row) => row.referral_inviter !== null)
        .map((row) => [String(row.months), row.referral_inviter])
    );
    tariff.referral_bonus_days_referee = Object.fromEntries(
      rows
        .filter((row) => row.referral_referee !== null)
        .map((row) => [String(row.months), row.referral_referee])
    );
    const topupPackages = packageSetFromRows(draft.topupRows, "gb", defaultCurrency);
    if (topupPackages) tariff.topup_packages = topupPackages;
  } else {
    const trafficPackages = packageSetFromRows(draft.trafficRows, "gb", defaultCurrency);
    if (trafficPackages) tariff.traffic_packages = trafficPackages;
    const conversion = parseNumber(draft.conversion_rate_rub_per_gb);
    if (conversion !== null) tariff.conversion_rate_rub_per_gb = conversion;
  }

  return tariff;
}
