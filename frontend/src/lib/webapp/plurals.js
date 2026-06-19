export function unitPluralBucket(value, lang) {
  void lang;
  return Number(value) === 1 ? "one" : "many";
}
