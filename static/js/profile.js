document.addEventListener("DOMContentLoaded", function () {
    const profileHTML = (username) => `
      <h3>Welcome to your profile page ${username}</h3>
      <div class="profile-container">
          <img src="static/assets/img/${username}/profileimage.png" alt="Photo de profil" class="profile-image">
      </div>
      <form id="uploadForm">
          <input id="username" name="username" value="${username}" type="hidden">
          <input type="file" name="image" id="imageInput" accept="image/*" required>
          <img id="preview" alt="Aperçu de l'image" style="max-width: 200px; display: none;">
          <button type="submit">Envoyer</button>
      </form>
      <p id="responseMessage"></p>
      <a href="/profile?update=true" class="button">Update Profile</a>
    `;
  
    const profileContainer = document.getElementById("profile-container");
  
    if (profileContainer) {
      const username = profileContainer.getAttribute("data-username") || "User";
      profileContainer.innerHTML = profileHTML(username);
  
      const imageInput = document.getElementById("imageInput");
      const preview = document.getElementById("preview");
      const uploadForm = document.getElementById("uploadForm");
      const responseMessage = document.getElementById("responseMessage");
  
      // Afficher l'aperçu de l'image avant l'envoi
      imageInput.addEventListener("change", function (event) {
        const file = event.target.files[0];
        if (file) {
          const reader = new FileReader();
          reader.onload = function (e) {
            preview.src = e.target.result;
            preview.style.display = "block";
          };
          reader.readAsDataURL(file);
        }
      });
  
      // Envoyer l'image en AJAX
      uploadForm.addEventListener("submit", function (event) {
        event.preventDefault();
  
        const formData = new FormData(uploadForm);
  
        fetch("/upload-profile-image", {
          method: "POST",
          body: formData,
        })
          .then((response) => response.json())
          .then((data) => {
            if (data.success) {
              responseMessage.textContent = "Image uploaded successfully!";
              responseMessage.style.color = "green";
            } else {
              responseMessage.textContent = "Error uploading image!";
              responseMessage.style.color = "red";
            }
          })
          .catch((error) => {
            console.error("Error:", error);
            responseMessage.textContent = "An error occurred!";
            responseMessage.style.color = "red";
          });
      });
    } else {
      console.error("La div #profile-container n'existe pas !");
    }
  });
  