document.addEventListener("DOMContentLoaded", function () {
  const footerHTML = `
<footer>
    <p>&copy; 2024 {{.Footer.SiteName}}. All rights reserved.</p>
</footer>`;

  const footersection = document.getElementById("footer");
  footersection.innerHTML = footerHTML;
});
