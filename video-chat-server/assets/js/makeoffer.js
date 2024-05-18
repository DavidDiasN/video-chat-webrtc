

function submitOfferName() {
  let offerName = document.getElementById("user-input").value;

  fetch("http://127.0.0.1:4009/videocall/offernameValidation", {
    method: "POST",
    body: offerName,
    headers: {
      "access-control-allow-origin": "*",
      "access-control-allow-methods":  "POST, GET, OPTIONS, PUT, DELETE",
      "access-control-allow-headers": "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token",
      "access-control-expose-headers":  "Authorization"
    } 
  })
  .then((response) => response.text())
  .then((text) => console.log(text));


}


