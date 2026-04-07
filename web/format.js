export function humanReadableAmount(value) {
  const s = String(value).trim().replace(/,/g, "");
  if (s === "") return "";
  const n = parseFloat(s);
  if (Number.isNaN(n)) return value;
  const abs = Math.abs(n);
  const sign = n < 0 ? "-" : "";
  if (abs < 1000) {
    const formatted = Number.isInteger(n) ? String(Math.trunc(n)) : n.toFixed(2).replace(/\.?0+$/, "");
    return formatted;
  }
  if (abs < 1e6) {
    const x = abs / 1000;
    const rounded = Math.round(x * 100) / 100;
    const str = rounded === Math.trunc(rounded) ? String(Math.trunc(rounded)) : rounded.toFixed(2).replace(/\.?0+$/, "");
    return sign + str + "k";
  }
  const x = abs / 1e6;
  const rounded = Math.round(x * 100) / 100;
  const str = rounded === Math.trunc(rounded) ? String(Math.trunc(rounded)) : rounded.toFixed(2).replace(/\.?0+$/, "");
  return sign + str + "M";
}

export function isMoneyColumn(type) {
  return type === "money_ars" || type === "money_usd";
}

export function isShareColumn(type) {
  return type === "share";
}
