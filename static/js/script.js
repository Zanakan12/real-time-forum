document.addEventListener("DOMContentLoaded", async () => {
  let socket;
  let username;
  const sendMessageButton = document.getElementById("send-msg-button");
  sendMessageButton.addEventListener("click", () => {
    sendMessage();
  });
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

  async function fetchMessages(action) {
    try {
      const response = await fetch("https://localhost:8080/api/chat");

      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      let messages = await response.json();

      if (!Array.isArray(messages) || messages.length === 0) {
        console.warn("⚠️ Aucun message disponible.");
        return;
      }
      console.log(action)
      if (action) {
        const lastMessage = messages.at(-1);
        if (lastMessage) {
          appendMessage(
            lastMessage.username,
            lastMessage.content,
            lastMessage.created_at
          );
        }
      } else {
        console.log(action);
        messages.forEach((msg) => {
          appendMessage(msg.username, msg.content, msg.created_at);
        });
      }
    } catch (error) {
      console.error("❌ Erreur lors de la récupération des messages :", error);
    }
  }

  async function fetchConnectedUsers() {
    try {
      const response = await fetch("https://localhost:8080/api/users");
      const users = await response.json();
      updateUserList(users);
    } catch (error) {
      console.error(
        "❌ Erreur lors de la récupération des utilisateurs connectés :",
        error
      );
    }
  }

  function connectWebSocket() {
    socket = new WebSocket("wss://localhost:8080/ws");

    socket.onopen = function () {
      console.log("✅ Connexion WebSocket établie !");
      fetchConnectedUsers(); // Met à jour la liste des utilisateurs connectés
    };

    socket.onmessage = function (event) {
      try {
        const msg = JSON.parse(event.data);
        console.log("📩 Message reçu :", msg);

        if (msg.type === "user_list") {
          try {
            const userList = JSON.parse(msg.content);
            if (Array.isArray(userList)) {
              updateUserList(userList);
            } else {
              console.error(
                "❌ Erreur : `user_list` n'est pas un tableau valide :",
                userList
              );
            }
          } catch (error) {
            console.error(
              "❌ Erreur lors du parsing de `user_list` :",
              error,
              msg.content
            );
          }
        } else if (msg.type === "message") {
          appendMessage(msg.username, msg.content);
        }
      } catch (error) {
        console.error(
          "❌ Erreur lors du parsing du message WebSocket :",
          error,
          event.data
        );
      }
    };

    socket.onclose = function () {
      console.warn("⚠️ Connexion WebSocket fermée.");
    };
  }

  function sendMessage(recipient) {
    const messageInput = document.getElementById("message");
    const message = messageInput.value.trim();
    if (message && socket.readyState === WebSocket.OPEN) {
      socket.send(
        JSON.stringify({
          type: "message",
          username: username,
          recipient: "voyou",
          content: message,
        })
      );
      messageInput.value = "";
    }
    fetchMessages(true);
  }

  function updateUserList(users) {
    console.log("👥 Mise à jour de la liste des utilisateurs :", users);
    const usersList = document.getElementById("users");
    usersList.innerHTML = "";

    JSON.parse(users).forEach((user) => {
      const li = document.createElement("li");
      li.textContent = user;
      usersList.appendChild(li);
    });
  }

  function appendMessage(username, content, createa_at) {
    const messagesList = document.getElementById("messages");
    const li = document.createElement("li");
    li.textContent = `${username}: ${content} ${createa_at}`;
    messagesList.appendChild(li);
  }

  console.log("🚀 - Page chargée !");
  await fetchUserData();
  await fetchMessages(false);
});
