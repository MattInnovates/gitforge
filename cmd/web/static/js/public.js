(function () {
  const ownerInput = document.getElementById("ownerInput");
  if (!ownerInput) {
    return;
  }

  ownerInput.addEventListener("keydown", function (event) {
    if (event.key === "Escape") {
      ownerInput.value = "";
    }
  });

  const cloneButtons = document.querySelectorAll(".copy-clone");
  cloneButtons.forEach(function (button) {
    button.addEventListener("click", async function () {
      const cloneURL = button.getAttribute("data-clone-url");
      if (!cloneURL) {
        return;
      }

      try {
        await navigator.clipboard.writeText(cloneURL);
        const original = button.textContent;
        button.textContent = "Copied";
        setTimeout(function () {
          button.textContent = original;
        }, 1200);
      } catch (_error) {
        alert("Copy failed. URL: " + cloneURL);
      }
    });
  });
})();
