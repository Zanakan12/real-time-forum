document.addEventListener("DOMContentLoaded", async () => {
  let socket;
  let username;
  let recipientSelect;
  let onlineUser;
  let offlineUser;

  const openChatBtn = document.getElementById("open-chat");

  // Fonction pour ouvrir la liste
  function open(arg) {
    const element = document.getElementById(arg);
    if (element.classList.contains("hidden")) {
      element.classList.remove("hidden"); // Ouvre la liste
      if (element.classList.contains("all-users")) fetchUserData();
      if (element.classList.contains("all-users")) fetchAllUsers();
      if (element.classList.contains("chat")) fetchMessages(recipientSelect);

      console.log("Téléchargement des statuts des utilisateurs terminé !");
    } else {
      element.classList.add("hidden"); // Ferme la liste
    }
  }

  // Fonction pour fermer la liste
  function close(arg) {
    const element = document.getElementById(arg);
    element.classList.add("hidden");
  }

  // Gérer l'ouverture du chat
  openChatBtn.addEventListener("click", (event) => {
    event.stopPropagation(); // Empêche la propagation pour éviter la fermeture immédiate
    const element = document.getElementById("all-users");
    open("all-users");
    // Gérer la fermeture du chat en cliquant à l'extérieur
    document.addEventListener("click", (event) => {
      if (!element.contains(event.target) && event.target !== openChatBtn) {
        close("all-users");
      }
    });
  });

  document
    .getElementById("users-online")
    .addEventListener("click", handleUserSelection);

  document
    .getElementById("users-offline")
    .addEventListener("click", handleUserSelection);

  function handleUserSelection(event) {
    if (event.target.classList.contains("selectUser")) {
      recipientSelect = event.target.id;
      let isOnline = event.target.classList.contains("online");

      console.log(
        `Utilisateur sélectionné : ${recipientSelect}, En ligne : ${isOnline}`
      );

      // Envoyer l'ID au backend Go
      fetch(`/api/chat?recipient=${recipientSelect}`).catch((error) =>
        console.error("Erreur lors de la récupération des messages :", error)
      );

      fetchMessages(recipientSelect);
      open("chat");
      close("all-users");
    }
  }

  const messageInput = document.getElementById("message");
  document
    .getElementById("message")
    .addEventListener("keydown", function (event) {
      if (event.key === "Enter") {
        document.getElementById("send-msg-button").click();
      }
    });

  document.getElementById("messages").addEventListener("scroll", function () {
    if (this.scrollTop === 0) {
      //loadOlderMessages(); // Fonction pour récupérer les anciens messages
    }
  });

  /*function loadOlderMessages() {
    const messagesList = document.getElementById("messages");

    for (let i = 0; i < 5; i++) {
      // Simulation de chargement de 5 anciens messages
      let oldMessage = document.createElement("li");
      oldMessage.textContent = "Ancien message " + (i + 1);
      oldMessage.classList.add("received");
      messagesList.prepend(oldMessage);
    }
  }*/

  const sendMessageButton = document.getElementById("send-msg-button");
  sendMessageButton.addEventListener("click", () => sendMessage());

  // Récupérer les infos utilisateur
  async function fetchUserData() {
    try {
      const response = await fetch("https://localhost:8080/api/get-user");
      const data = await response.json();
      if (data.username) {
        username = data.username;
        connectWebSocket();
      } else {
        window.location.href = "/login";
      }
    } catch (error) {
      console.error(
        "❌ Erreur lors de la récupération de l'utilisateur :",
        error
      );
      window.location.href = "/login";
    }
  }

  // Récupérer les anciens messages
  async function fetchMessages(recipientSelect) {
    if (recipientSelect === undefined) return;
    try {
      const response = await fetch(
        `https://localhost:8080/api/chat?recipient=${recipientSelect}`
      );
      if (!response.ok)
        throw new Error(`HTTP error! Status: ${response.status}`);

      let messages = await response.json();
      messages = JSON.parse(messages);

      if (!Array.isArray(messages)) {
        return console.warn("⚠️ Aucun message disponible.");
      }
      const messagesList = document.getElementById("messages");
      messagesList.innerHTML = "";
      messages.forEach((msg) => {
        let isSender = false;
        if (msg.username === username) {
          isSender = true;
        }
        appendMessage(
          msg.username,
          msg.recipient,
          msg.content,
          msg.created_at,
          isSender
        );
      });
    } catch (error) {
      console.error("❌ Erreur lors de la récupération des messages :", error);
    }
  }

  // Récupérer la liste des utilisateurs connectés
  async function fetchConnectedUsers() {
    try {
      const response = await fetch(
        "https://localhost:8080/api/users-connected"
      );
      const users = await response.json();
      updateUserList(JSON.parse(users));
    } catch (error) {
      console.error(
        "❌ Erreur lors de la récupération des utilisateurs connectés :",
        error
      );
    }
  }

  // Connexion WebSocket
  function connectWebSocket() {
    socket = new WebSocket(`wss://localhost:8080/ws?username=${username}`);

    socket.onopen = () => {
      console.log("✅ Connexion WebSocket établie !");
      fetchConnectedUsers();
    };

    socket.onclose = () => console.warn("⚠️ Connexion WebSocket fermée.");
  }

  // Envoi de message
  function sendMessage() {
    const recipient = recipientSelect;
    const message = messageInput.value.trim();
    if (!recipient || !message) {
      alert("Veuillez entrer un destinataire et un message !");
      return;
    }

    if (socket.readyState === WebSocket.OPEN) {
      const msgObj = {
        type: "message",
        username: username,
        recipient: recipient,
        content: message,
      };

      socket.send(JSON.stringify(msgObj));
      appendMessage(
        username,
        recipient,
        message,
        new Date().toISOString(),
        true
      ); // Affichage immédiat
      console.log("message envoyé");
      messageInput.value = "";
    } else {
      alert("WebSocket non connecté !");
    }
  }

  // Mettre à jour la liste des utilisateurs connectés
  function updateUserList(users) {
    console.log("👥 Mise à jour de la liste des utilisateurs :", users);
    const usersList = document.getElementById("users-online");
    usersList.innerHTML = "";

    users.forEach((user) => {
      if (user !== username) {
        const li = document.createElement("li");
        li.textContent = user[0];
        li.classList.add("selectUser", "online");
        li.id = `${user}`;
        usersList.appendChild(li);
      }
    });
  }

  // Ajouter un message dans le chat
  function appendMessage(username, recipient, content, createdAt, isSender) {
    const messagesList = document.getElementById("messages");
    const li = document.createElement("li");

    li.classList.add("message");
    if (isSender) {
      li.classList.add("sent");
    } else {
      li.classList.add("received");
    }

    li.innerHTML = `${content} <small>${new Date(
      createdAt
    ).toLocaleTimeString()}</small>`;
    messagesList.appendChild(li);

    // Vérifier si l'utilisateur est en bas avant de scroller
    let isScrolledToBottom =
      messagesList.scrollHeight - messagesList.clientHeight <=
      messagesList.scrollTop + 1;

    if (isScrolledToBottom) {
      messagesList.scrollTop = messagesList.scrollHeight; // Scroll en bas seulement si l'utilisateur est déjà en bas
    }
  }

  async function fetchAllUsers() {
    try {
      const response = await fetch("https://localhost:8080/api/all-user");
      if (!response.ok) {
        throw new Error("Erreur lors de la récupération des utilisateurs");
      }
      const users = await response.json();

      const filtredUser = users.sort((a, b) =>
        a.Username.localeCompare(b.Username)
      );
      // Affichage sur la page HTML (si nécessaire)
      const userList = document.getElementById("users-offline");
      filtredUser.forEach((user) => {
        if (user !== username) {
          const li = document.createElement("li");
          li.textContent = user.Username[0].toUpperCase();
          li.classList.add("selectUser", "offline", "short");
          li.id = `${user.Username}`;
          userList.appendChild(li);
        }
      });
    } catch (error) {
      console.error("Erreur :", error);
    }
  }

  console.log("🚀 - Page chargée !");
  await fetchUserData();
  await fetchAllUsers();
});

/*
btnProfile.style.backgroundImage = `url('static/assets/img/${username}/profileimage.png')`;
  btnProfile.style.backgroundSize = "cover"; // Ajuste l'image
  btnProfile.style.backgroundPosition = "center"; // Centre l'image
  btnProfile.style.backgroundRepeat = "no-repeat"; // Empêche la répétition
  
const btnProfile = document.getElementById(profile - image - nav);
  console.log(btnProfile.textContent);
  console.log(`url('static/assets/img/${username}/profileimage.png')`);

document
  .getElementById("imageInput")
  .addEventListener("change", function (event) {
    console.log("telechargement en CountQueuingStrategy");
    const file = event.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = function (e) {
        const preview = document.getElementById("preview");
        preview.src = e.target.result;
        preview.style.display = "block";
      };
      reader.readAsDataURL(file);
    }
  });

document
  .getElementById("uploadForm")
  .addEventListener("submit", async function (event) {
    event.preventDefault();

    const formData = new FormData();
    formData.append(
      "user-profile",
      document.getElementById("user-profile").value
    );
    formData.append("image", document.getElementById("imageInput").files[0]);

    const response = await fetch("http://localhost:8080/upload", {
      method: "POST",
      body: formData,
    });

    const result = await response.text();
    document.getElementById("responseMessage").innerText = result;
  });*/
