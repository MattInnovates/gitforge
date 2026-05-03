(function () {
  const ownerInput = document.getElementById("adminOwnerInput");
  if (!ownerInput) {
    return;
  }

  const storageKey = "gitforge:lastOwner";
  const remembered = localStorage.getItem(storageKey);
  if (!ownerInput.value && remembered) {
    ownerInput.value = remembered;
  }

  ownerInput.addEventListener("change", function () {
    if (ownerInput.value.trim() !== "") {
      localStorage.setItem(storageKey, ownerInput.value.trim());
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
        }, 1000);
      } catch (_error) {
        alert("Copy failed. URL: " + cloneURL);
      }
    });
  });
})();
