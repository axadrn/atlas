(function () {
  const roots = () => document.querySelectorAll("[data-place-search]");

  function setup(root) {
    if (root.dataset.placeSearchReady === "true") return;
    const input = root.querySelector("[data-place-search-input]");
    if (!input) return;
    root.dataset.placeSearchReady = "true";

    let debounceTimer;
    let requestController;

    function elements() {
      const content = document.getElementById(input.getAttribute("aria-controls"));
      if (!content) return {};
      return {
        content,
        empty: content.querySelector("[data-place-search-empty]"),
        list: content.querySelector("[data-place-search-results]"),
        popup: content.querySelector("[data-tui-combobox-popup]"),
      };
    }

    function setEmpty(message) {
      const { empty, list, popup } = elements();
      if (!empty || !list || !popup) return;
      empty.textContent = message;
      list.replaceChildren();
      list.toggleAttribute("data-empty", true);
      popup.toggleAttribute("data-empty", true);
    }

    function render(html) {
      const { list, popup } = elements();
      if (!list || !popup) return;
      const template = document.createElement("template");
      template.innerHTML = html;
      list.replaceChildren(template.content);
      const hasResults = list.querySelector("[data-tui-combobox-item]") !== null;
      list.toggleAttribute("data-empty", !hasResults);
      popup.toggleAttribute("data-empty", !hasResults);
      if (!hasResults) {
        const { empty } = elements();
        if (empty) empty.textContent = "No matching place yet.";
      }
    }

    async function search(query) {
      if (requestController) requestController.abort();
      requestController = new AbortController();
      try {
        const response = await fetch("/fragments/place-search?q=" + encodeURIComponent(query), {
          signal: requestController.signal,
          headers: { Accept: "text/html" },
        });
        if (!response.ok) throw new Error("search failed");
        const html = await response.text();
        if (input.value.trim() === query) render(html);
      } catch (error) {
        if (error.name !== "AbortError") setEmpty("Search is temporarily unavailable.");
      }
    }

    input.addEventListener("input", () => {
      clearTimeout(debounceTimer);
      const query = input.value.trim();
      if (query.length < 2) {
        if (requestController) requestController.abort();
        setEmpty("Type at least two characters.");
        return;
      }
      setEmpty("Searching...");
      debounceTimer = setTimeout(() => search(query), 140);
    });

    root.addEventListener("combobox-change", (event) => {
      const slug = event.detail && event.detail.values && event.detail.values[0];
      if (slug) window.location.assign("/places/" + encodeURIComponent(slug));
    });
  }

  function init() {
    roots().forEach(setup);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
  new MutationObserver(init).observe(document.body, { childList: true, subtree: true });

  document.addEventListener("keydown", (event) => {
    if (event.key !== "/") return;
    const input = document.querySelector("[data-place-search-input]");
    if (!input || document.activeElement === input) return;
    event.preventDefault();
    input.focus();
  });
})();
