addEventListener("DOMContentLoaded", function () {
  const chatMessages = document.getElementById("chat-messages");
  const unfoldButton = document.getElementById("item");

  unfoldButton.innerHTML = `
  <svg width="30" height="30" viewBox="0 0 24 24" fill="white" xmlns="http://www.w3.org/2000/svg">
      <path d="M8 10L12 14L16 10" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
  </svg>
`;

  unfoldButton.addEventListener("click", function () {
    if (chatMessages.classList.contains("unfold")) {
      chatMessages.classList.remove("unfold");
      chatMessages.classList.add("fold");
    } else {
      chatMessages.classList.remove("fold");
      chatMessages.classList.add("unfold");
    }
  });
});
