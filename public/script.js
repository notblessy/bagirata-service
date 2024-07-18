document.addEventListener("DOMContentLoaded", () => {
  const copyButton = document.getElementById("copyButton");
  const alertBox = document.getElementById("alert");
  const alertDismissButton = document.getElementById("alertDismiss");
  const bankNumber = document.getElementById("bankNumber").innerText;

  function showAlert() {
    alertBox.classList.remove("hidden");
    setTimeout(() => {
      hideAlert();
    }, 5000);
  }

  function hideAlert() {
    alertBox.classList.add("hidden");
  }

  copyButton.addEventListener("click", () => {
    navigator.clipboard.writeText(bankNumber);
    showAlert();
  });

  alertDismissButton.addEventListener("click", () => {
    hideAlert();
  });
});
