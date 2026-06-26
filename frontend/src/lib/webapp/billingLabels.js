/**
 * Returns the default payment method ID from the available methods list.
 * @param {Array} methods
 * @returns {string}
 */
export function defaultPaymentMethod(methods) {
  if (!Array.isArray(methods) || !methods.length) return "";
  return String(methods[0]?.id || methods[0] || "");
}

/**
 * Returns the primary CTA label for the home screen pay/subscribe button.
 * @param {object} params
 * @param {boolean} params.subscriptionActive
 * @param {boolean} params.trafficMode
 * @param {object} params.appSettings
 * @param {object} params.selectedPlan
 * @param {(key: string, params?: object, fallback?: string) => string} params.t
 * @returns {string}
 */
export function primaryPayActionLabel({ subscriptionActive, trafficMode, appSettings, selectedPlan, t }) {
  if (!subscriptionActive && appSettings?.trial_enabled && appSettings?.trial_available) {
    return t("wa_pay_full_subscription", {}, "Pay for full subscription");
  }
  if (trafficMode || selectedPlan?.sale_mode === "traffic_package") return t("wa_buy_traffic");
  return subscriptionActive
    ? t("wa_renew_subscription", {}, "Renew subscription")
    : t("wa_pay_subscription");
}
