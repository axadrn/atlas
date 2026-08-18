(function () {
  let installPrompt;
  const installButton = document.querySelector("[data-pwa-install]");
  const helpTrigger = document.querySelector("[data-pwa-help-trigger]");

  function installed() {
    return window.matchMedia("(display-mode: standalone)").matches || window.navigator.standalone === true;
  }

  function updateButton() {
    if (installButton) installButton.hidden = installed();
  }

  if ("serviceWorker" in navigator) {
    window.addEventListener("load", function () {
      navigator.serviceWorker.register("/service-worker.js").catch(function () {
        // Atlas remains a normal website when service workers are unavailable.
      });
    });
  }

  if (!installButton) return;

  window.addEventListener("beforeinstallprompt", function (event) {
    event.preventDefault();
    installPrompt = event;
    updateButton();
  });

  window.addEventListener("appinstalled", function () {
    installPrompt = undefined;
    updateButton();
  });

  installButton.addEventListener("click", async function () {
    if (!installPrompt) {
      if (helpTrigger) helpTrigger.click();
      return;
    }
    await installPrompt.prompt();
    await installPrompt.userChoice;
    installPrompt = undefined;
    updateButton();
  });

  updateButton();
})();
