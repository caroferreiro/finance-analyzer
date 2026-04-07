(function (global) {
  "use strict";

  var documentRef = global.document;
  if (!documentRef) {
    return;
  }
  if (global.__foFinanceOverviewReady && typeof global.__foFinanceOverviewReady.then === "function") {
    return;
  }

  var assetHelpers = global.__FO_ASSET_HELPERS__ || {};

  var loaderSrc =
    documentRef.currentScript && documentRef.currentScript.src ? documentRef.currentScript.src : documentRef.baseURI;
  var overviewUrl = new URL("finance-overview.js", loaderSrc).href;
  var token =
    typeof assetHelpers.readCanonicalVersionToken === "function"
      ? String(assetHelpers.readCanonicalVersionToken({ scriptHref: loaderSrc }) || "")
      : String(global.__FO_ASSET_VERSION__ || "");
  var versionedOverviewUrl =
    typeof assetHelpers.appendVersionParam === "function"
      ? assetHelpers.appendVersionParam(overviewUrl, token)
      : overviewUrl;

  global.__FO_FINANCE_OVERVIEW_VERSION__ = token;
  global.__FO_FINANCE_OVERVIEW_URL__ = versionedOverviewUrl;

  global.__foFinanceOverviewReady = new Promise(function (resolve, reject) {
    var script = documentRef.createElement("script");
    script.src = versionedOverviewUrl;
    script.setAttribute("data-fo-overview-loader", "true");
    script.onload = function () {
      resolve(global.FinanceOverview);
    };
    script.onerror = function () {
      reject(new Error("Failed to load finance-overview runtime: " + versionedOverviewUrl));
    };
    documentRef.head.appendChild(script);
  });
})(window);
