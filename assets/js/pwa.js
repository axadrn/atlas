(function () {
  let installPrompt;
  const installButtons = document.querySelectorAll("[data-pwa-install]");
  const helpTrigger = document.querySelector("[data-pwa-help-trigger]");

  function installed() {
    return window.matchMedia("(display-mode: standalone)").matches || window.navigator.standalone === true;
  }

  function updateButton() {
    installButtons.forEach(function (button) {
      button.hidden = installed();
    });
  }

  if ("serviceWorker" in navigator) {
    window.addEventListener("load", function () {
      navigator.serviceWorker.register("/service-worker.js").catch(function () {
        // Atlas remains a normal website when service workers are unavailable.
      });
    });
  }

  if (installButtons.length === 0) return;

  window.addEventListener("beforeinstallprompt", function (event) {
    event.preventDefault();
    installPrompt = event;
    updateButton();
  });

  window.addEventListener("appinstalled", function () {
    installPrompt = undefined;
    updateButton();
  });

  installButtons.forEach(function (button) {
    button.addEventListener("click", async function () {
      if (!installPrompt) {
        if (helpTrigger) window.setTimeout(function () { helpTrigger.click(); }, 0);
        return;
      }
      await installPrompt.prompt();
      await installPrompt.userChoice;
      installPrompt = undefined;
      updateButton();
    });
  });

  updateButton();
})();
