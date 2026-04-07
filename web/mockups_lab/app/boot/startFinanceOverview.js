export async function startFinanceOverview(options = {}) {
  const {
    variant = "mockup1",
    scope = window,
    doc = document,
    bootConfig = {},
  } = options;

  try {
    await scope.__foFinanceOverviewReady;
    scope.FinanceOverview.boot({ variant, ...bootConfig });
  } catch (err) {
    const message = err && err.message ? err.message : String(err);
    if (
      scope.FinanceOverview &&
      scope.FinanceOverview.helpers &&
      typeof scope.FinanceOverview.helpers.showBlockingError === "function"
    ) {
      scope.FinanceOverview.helpers.showBlockingError(message);
      return;
    }
    const fallback = doc.createElement("pre");
    fallback.textContent = "Blocking startup error\n" + message;
    fallback.style.whiteSpace = "pre-wrap";
    fallback.style.padding = "16px";
    fallback.style.color = "#f3f4f6";
    fallback.style.background = "#111827";
    fallback.style.border = "1px solid #374151";
    fallback.style.margin = "12px";
    doc.body.appendChild(fallback);
  }
}
