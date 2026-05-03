(function () {
  const cloneButtons = document.querySelectorAll('.copy-clone');

  const setCopiedState = function (button) {
    const originalText = button.textContent;
    button.textContent = 'Copied';
    button.disabled = true;

    window.setTimeout(function () {
      button.textContent = originalText;
      button.disabled = false;
    }, 1200);
  };

  cloneButtons.forEach(function (button) {
    button.addEventListener('click', async function () {
      const cloneURL = button.getAttribute('data-clone-url');
      if (!cloneURL) {
        return;
      }

      try {
        await navigator.clipboard.writeText(cloneURL);
        setCopiedState(button);
      } catch (_err) {
        window.prompt('Copy this URL manually:', cloneURL);
      }
    });
  });
})();