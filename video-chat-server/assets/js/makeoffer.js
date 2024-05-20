function submitOfferName() {
  let offerName = document.getElementById("user-input").value;

  fetch("http://127.0.0.1:4009/videocall/offernameValidation", {
    method: "POST",
    body: offerName,
  })
  .then((response) => response.text())
  .then((text) => console.log(text));


}
