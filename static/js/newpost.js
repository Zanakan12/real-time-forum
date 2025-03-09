document.addEventListener("DOMContentLoaded", function () {
    const newPostHTML = `
      <form action="/post-validation" method="post" enctype="multipart/form-data">
        <table>
          <div id="error-messages"></div>
          <div id="categories-container"></div>
          <tr>
            <td colspan="4"><hr width="100%"></td>
          </tr>
          <tr>
            <td><label for="body">Post content:</label></td>
            <td colspan="2"><input id="body" name="body" type="text" required/></td>
          </tr>
          <tr>
            <td><label for="image">Select an image:</label></td>
            <td colspan="2"><input id="image" name="image" type="file" accept="image/*"/></td>
          </tr>
          <tr>
            <td><input type="submit" value="Submit" /></td>
          </tr>
        </table>
      </form>
    `;
  
    const newPostContainer = document.getElementById("newpost-container");
    if (newPostContainer) {
      newPostContainer.innerHTML = newPostHTML;
    } else {
      console.error("La div #newpost-container n'existe pas !");
    }
  });
  