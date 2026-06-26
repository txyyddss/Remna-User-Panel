/**
 * Maps a trial activation error to a user-facing message.
 * @param {object} error
 * @param {(key: string, params?: object, fallback?: string) => string} t
 * @returns {string}
 */
export function trialActivationFailureMessage(error, t) {
  if (
    error?.error === "trial_telegram_required" ||
    error?.message === "telegram_required" ||
    error?.message === "disposable_email"
  ) {
    return t(
      "wa_trial_telegram_required_error",
      {},
      "Link Telegram to activate the trial."
    );
  }
  return error?.message || t("wa_trial_activation_failed");
}

/**
 * Maps a referral welcome bonus error to a user-facing message.
 * @param {object} error
 * @param {(key: string, params?: object, fallback?: string) => string} t
 * @returns {string}
 */
export function referralWelcomeFailureMessage(error, t) {
  if (
    error?.error === "referral_welcome_telegram_required" ||
    error?.message === "telegram_required" ||
    error?.message === "disposable_email"
  ) {
    return t(
      "wa_referral_welcome_telegram_required_error",
      {},
      "Link Telegram to get the referral bonus."
    );
  }
  return error?.message || t("wa_referral_welcome_claim_failed");
}
