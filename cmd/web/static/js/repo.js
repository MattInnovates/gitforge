(function () {
  const cloneButtons = document.querySelectorAll(".copy-clone");
  cloneButtons.forEach(function (button) {
    button.addEventListener("click", async function () {
      const cloneURL = button.getAttribute("data-clone-url");
      if (!cloneURL) {
        return;
      }

      try {
        await navigator.clipboard.writeText(cloneURL);
        const originalText = button.textContent;
        button.textContent = "Copied";
        setTimeout(function () {
          button.textContent = originalText;
        }, 1000);
      } catch (_err) {
        alert("Copy failed. URL: " + cloneURL);
      }
    });
  });
})();
