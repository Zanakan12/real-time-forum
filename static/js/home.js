document.getElementById("content").innerHTML = `
    <h4>{{.UserUsername}}, {{if eq .UserRole "traveler"}}register to {{end}}tell us a story...</h4>
    {{if ne .UserRole "traveler"}}
    <div id="newpost-container"></div>
    {{end}}
    <div id="categories-selection-container"></div>
    <div id="lastposts-container"></div>
    <div id="chat-messages" class="fold">
        <section>
            <div id="all-users" class="hidden">
                <h3>Utilisateurs connectés :</h3>
                <ul id="users-online" name="user"></ul>
                <h3>Utilisateurs hors ligne :</h3>
                <ul id="users-offline"></ul>
            </div>

            <div id="chat" class="hidden">
                <div id="users-logo"> </div>
                <ul id="messages"></ul>

                <div id="chat-input-container">
                    <input id="message" type="text" placeholder="Écrivez un message">
                    <input id="send-msg-button" type="button" value="S">
                </div>
            </div>
        </section>
    </div>
`;
