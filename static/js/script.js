document.addEventListener("DOMContentLoaded", async () => {
  let socket;
  let username;
  let recipientSelect;
  let onlineUser;
  let offlineUser;
  // Sélection des éléments HTML
  const sendMessageButton = document.getElementById("send-msg-button");
  const messageInput = document.getElementById("message");

  async function getUserSelected(username) {
    console.log("username", username);
    try {
      await fetch(`/api/chat?recipient=${username}`);
      fetchMessages(username);
    } catch (error) {
      console.error("Error", error);
    }
  }

  document
    .getElementById("message")
    .addEventListener("keydown", function (event) {
      if (event.key === "Enter") {
        document.getElementById("send-msg-button").click();
      }
    });
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
    console.log("longueur", recipientSelect);
    if (recipientSelect === undefined) return;
    try {
      const response = await fetch(
        `https://localhost:8080/api/chat?recipient=${recipientSelect}`
      );
      if (!response.ok)
        throw new Error(`HTTP error! Status: ${response.status}`);

      let messages = await response.json();
      messages = JSON.parse(messages);

      if (!Array.isArray(messages))
        return console.warn("⚠️ Aucun message disponible.");

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

    socket.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        console.log("📩 Message reçu :", msg);

        if (msg.type === "user_list") {
          updateUserList(JSON.parse(msg.content));
        } else if (msg.type === "message") {
          appendMessage(
            msg.username,
            msg.recipient,
            msg.content,
            msg.created_at
          );
        }
      } catch (error) {
        console.error(
          "❌ Erreur lors du parsing du message WebSocket :",
          error,
          event.data
        );
      }
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
        usersList.addEventListener("click", () => getUserSelected(user));
        li.classList.add("selectUser", "online");
        li.id = `${user}`;
        usersList.appendChild(li);
      }
    });
  }

  // Ajouter un message dans le chat
  function appendMessage(username, recipient, content, createdAt, isSender) {
    console.log(recipientSelect);
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
      console.log(users);
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
  await fetchMessages();
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
