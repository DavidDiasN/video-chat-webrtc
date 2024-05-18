document.getElementById("submit-offer-name").addEventListener("click", submitOffer);
 
let publicServer = "http://127.0.0.1:4009"

function submitOffer() {
  let input = document.getElementById("user-input").value;
  console.log(input)
  fetch(publicServer + "/videocall/getoffers") // replace with your server URL
    .then(response => response.json())
    .then((info) => {
      if (info.includes(input)) {
        return
      }
    })
    .catch(error => console.error('Error:', error));

  //if (catchData.includes(input)) {
  //  return
  //}
  console.log("made it past includes")

  fetch(publicServer + "/videocall/postoffer", {
  method: 'POST',
  body: input, // Send a JSON object with the input value
})
  .then(response => response.json())
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));  
  //JUST USE JS PLEASE PLEASE PLEASE
}
