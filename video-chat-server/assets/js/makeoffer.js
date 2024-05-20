function submitOfferName() {
  let inputButton = document.getElementById("user-input");
  let inputValue = inputButton.value;
  let hash = "";
  fetch("http://127.0.0.1:4009/videocall/offernameValidation", {
    method: "POST",
    body: inputValue,
  })
  .then((response) => response.text())
  .then((text) => {
    if (text === "NO") {
        console.log("it said no")
        return
        // have a popup say that name is taken
      } else {
        console.log(text);
        inputButton.disabled = true;
        hash = text;

        if (window['WebSocket']) {
          const conn = new WebSocket('ws://' + document.location.host + '/videocall/MakeOffer/ws');
          let initial = true;
          console.log("we got here") 
          conn.onopen = function () {
            conn.send(inputValue + " " + hash)
          }

          conn.onclose = () => {
            console.log("Connection closing")
          }


          conn.onmessage = evt => {
            if (initial) {
              console.log(evt.data)
            } else {
              console.log("wooo bye")
              conn.send("DONE")
              conn.close()
            }
            
          }


        }
      }

    });

  // start websocket connection and supply


}
