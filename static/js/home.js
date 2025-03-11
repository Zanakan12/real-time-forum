let username = "rafta";

document.addEventListener("DOMContentLoaded", function () {
  document.getElementById("home").innerHTML = `
      <h4> ${username}, tell us a story...</h4>
      
      <div id="newpost-container"></div>
      <div id="categories-selection-container"></div>
      <div id="lastposts-container"></div>
      <div id="chat-messages" class="fold">
          <section>
              <div id="all-users" class="hidden">
                  <h3>en ligne:</h3>
                  <ul id="users-online" name="user"></ul>
                  <h3>hors ligne :</h3>
                  <ul id="users-offline"></ul>
              </div>

              <div id="chat" class="hidden">
                  <div id="header-chat">
                    <div id="photo-chat"></div>
                    <div id="name-chat"></div>   
                    <div id="reduce-chat">_</div>
                    <div id="close-chat">x</div>
                  </div>
                  <ul id="messages"></ul>

                  <div id="chat-input-container">
                      <input id="message" type="text" placeholder="Écrivez un message">
                      <input id="send-msg-button" type="button">
                  </div>
              </div>
          </section>
      </div>
    `;
});
