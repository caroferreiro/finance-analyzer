import { createShellRouter } from "./router.js";

export function bindSidebarNavigation(doc = document) {
  const buttons = Array.from(doc.querySelectorAll(".nav button[data-page]"));
  const pages = Array.from(doc.querySelectorAll(".page"));
  const scrollContainer = doc.getElementById("main-content");
  const router = createShellRouter({
    buttons,
    pages,
    scrollContainer,
  });

  buttons.forEach((buttonNode) => {
    buttonNode.addEventListener("click", () => {
      router.show(buttonNode.dataset.page);
    });
  });

  return router;
}
