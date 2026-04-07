(function (global) {
  "use strict";

  // Canonical token for mockups shared asset generation.
  var ASSET_VERSION = "20260406-cfu001";

  function trimString(value) {
    return typeof value === "string" ? value.trim() : "";
  }

  function appendVersionParam(url, token) {
    var href = String(url || "");
    var cleanToken = trimString(token);
    if (!cleanToken) {
      return href;
    }
    var sep = href.indexOf("?") >= 0 ? "&" : "?";
    return href + sep + "v=" + encodeURIComponent(cleanToken);
  }

  function readVersionFromUrl(url, baseHref) {
    try {
      if (!url) {
        return "";
      }
      var parsed = new URL(
        String(url),
        baseHref || (global.location && global.location.href ? global.location.href : undefined)
      );
      return trimString(parsed.searchParams.get("v") || "");
    } catch {
      return "";
    }
  }

  function readCanonicalVersionToken(options) {
    var opts = options || {};
    var explicit = trimString(opts.explicitToken);
    if (explicit) {
      return explicit;
    }
    var staticToken = trimString(global.__FO_ASSET_VERSION__);
    if (staticToken) {
      return staticToken;
    }
    var overviewToken = trimString(global.__FO_FINANCE_OVERVIEW_VERSION__);
    if (overviewToken) {
      return overviewToken;
    }
    var scriptToken = readVersionFromUrl(opts.scriptHref);
    if (scriptToken) {
      return scriptToken;
    }
    try {
      if (!global.location || !global.location.href) {
        return "";
      }
      var locationUrl = new URL(global.location.href);
      return (
        trimString(locationUrl.searchParams.get("v") || "") ||
        trimString(locationUrl.searchParams.get("ts") || "")
      );
    } catch {
      return "";
    }
  }

  global.__FO_ASSET_VERSION__ = ASSET_VERSION;
  global.__FO_ASSET_HELPERS__ = {
    readCanonicalVersionToken: readCanonicalVersionToken,
    appendVersionParam: appendVersionParam,
    readVersionFromUrl: readVersionFromUrl,
  };
})(window);
