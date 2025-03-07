document.addEventListener("DOMContentLoaded", async () => {
  let socket;
  let username;
  let recipientSelect;
  // Sélection des éléments HTML
  const sendMessageButton = document.getElementById("send-msg-button");
  const messageInput = document.getElementById("message");

  document.getElementById("users").addEventListener("click", function (event) {
    if (event.target.classList.contains("selectUser")) {
      recipientSelect = event.target.textContent;
      // Envoyer l'ID au backend Go
      fetch(`/api/chat?recipient=${recipientSelect}`).catch((error) =>
        console.error("Erreur lors de la récupération des messages :", error)
      );
    }
    fetchMessages(recipientSelect);
  });

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
      const response = await fetch("https://localhost:8080/api/users");
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
    const usersList = document.getElementById("users");
    usersList.innerHTML = "";

    users.forEach((user) => {
      const li = document.createElement("li");
      li.textContent = user;
      li.classList.add("selectUser");
      li.id = `${user}`;
      usersList.appendChild(li);
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

    li.innerHTML = `<strong>${username} → ${recipient} :</strong> ${content} <small>(${new Date(
      createdAt
    ).toLocaleTimeString()})</small>`;
    messagesList.appendChild(li);
  }

  console.log("🚀 - Page chargée !");
  await fetchUserData();
  await fetchMessages();
});
