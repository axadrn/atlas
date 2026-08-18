(function () {
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
        empty: content.querySelector("[data-place-search-empty]"),
        list: content.querySelector("[data-place-search-results]"),
        popup: content.querySelector("[data-tui-combobox-popup]"),
      };
    }

    function showMessage(message) {
      const { empty, list, popup } = elements();
      if (!empty || !list || !popup) return;
      empty.textContent = message;
      list.replaceChildren();
      list.toggleAttribute("data-empty", true);
      popup.toggleAttribute("data-empty", true);
    }

    function showResults(html) {
      const { empty, list, popup } = elements();
      if (!empty || !list || !popup) return;

      const template = document.createElement("template");
      template.innerHTML = html;
      list.replaceChildren(template.content);

      const hasResults = list.querySelector("[data-tui-combobox-item]") !== null;
      list.toggleAttribute("data-empty", !hasResults);
      popup.toggleAttribute("data-empty", !hasResults);
      if (!hasResults) empty.textContent = "No matching place yet.";
    }

    // Queries under two characters resolve to the empty query, which the
    // server answers with the top curated destinations.
    function effectiveQuery() {
      const query = input.value.trim();
      return query.length < 2 ? "" : query;
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
        if (effectiveQuery() === query) showResults(html);
      } catch (error) {
        if (error.name !== "AbortError") showMessage("Search is temporarily unavailable.");
      }
    }

    input.addEventListener("input", function () {
      clearTimeout(debounceTimer);
      debounceTimer = setTimeout(function () {
        search(effectiveQuery());
      }, 140);
    });

    // Preload the top destinations so the popup is never empty.
    search("");

    root.addEventListener("combobox-change", function (event) {
      const slug = event.detail && event.detail.values && event.detail.values[0];
      if (slug) window.location.assign("/places/" + encodeURIComponent(slug));
    });
  }

  function init() {
    document.querySelectorAll("[data-place-search]").forEach(setup);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }

  document.addEventListener("keydown", function (event) {
	if (event.key !== "/") return;
	const inputs = Array.from(document.querySelectorAll("[data-place-search-input]"));
	const input = inputs.find(function (candidate) {
	  return candidate.getClientRects().length > 0;
	});
	if (input && document.activeElement === input) return;
	event.preventDefault();
	if (input) {
	  input.focus();
	  return;
	}
	document.querySelector("[data-place-search-trigger]")?.click();
  });
})();
