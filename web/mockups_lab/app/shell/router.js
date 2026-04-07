export function createShellRouter(options) {
  const opts = options || {};
  const buttons = Array.from(opts.buttons || []);
  const pages = Array.from(opts.pages || []);
  const scrollContainer = opts.scrollContainer || null;

  function show(pageId) {
    pages.forEach((pageNode) => {
      const isActive = pageNode && pageNode.id === pageId;
      if (!pageNode) {
        return;
      }
      pageNode.classList.toggle("active", isActive);
      pageNode.hidden = !isActive;
      pageNode.setAttribute("aria-hidden", isActive ? "false" : "true");
    });

    buttons.forEach((buttonNode) => {
      const isActive = buttonNode && buttonNode.dataset && buttonNode.dataset.page === pageId;
      if (!buttonNode) {
        return;
      }
      buttonNode.classList.toggle("active", isActive);
      buttonNode.setAttribute("aria-current", isActive ? "page" : "false");
    });

    if (scrollContainer) {
      scrollContainer.scrollTop = 0;
      scrollContainer.scrollLeft = 0;
    }
  }

  return {
    buttons,
    pages,
    scrollContainer,
    show,
  };
}
