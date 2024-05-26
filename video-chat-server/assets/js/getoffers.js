
const servers = {
  iceServers: [
    {
      urls: ['stun:stun1.l.google.com:19302', 'stun:stun2.l.google.com:19302']
    },
  ],
  iceCandidiatePoolSize: 10,
};
let conn;
let peerConn = new RTCPeerConnection(servers);
let localStream = null;
let remoteStream = null;

const webcamButton = document.getElementById("enable-webcam");
const webcamVideo = document.getElementById("local-video");
const remoteVideo = document.getElementById("remote-video");

// request to input a name and such. 
// check if the name is valid
// if valid establish websocket connection
// if invalid request the user try another username

function submitAnswerName() {
  let inputButton = document.getElementById("user-input");
  let inputValue = inputButton.value;
  let hash = "";
  fetch("http://127.0.0.1:4009/videocall/AnswerValidation", {
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
          conn = new WebSocket('ws://' + document.location.host + '/videocall/MakeAnswer/ws');
          let initial = true;

          let offerList = document.getElementById("offer-list")


          console.log("we got here") 

          conn.onopen = function () {
            conn.send(inputValue + " " + hash)
          };

          conn.onclose = () => {
            console.log("Connection closing")
          };

          conn.onmessage = evt => {
            console.log("MESSAGE RECV: " + evt.data);
            if (initial) {
              console.log(evt.data);
              initial = false;
            }            
                
            if (evt.data.slice(0,3) === "<li") {
              console.log("Made it here")
              // add it to the list
              offerList.innerHTML = evt.data
            } else {
              
              console.log("wooo bye")
              conn.send("DONE")
              conn.close()
            }

          };


        }
      }

    });
  // once the offer and answers have been exchanged we should end the websocket connection and just be in the call.
  //
}

async function clickName(name) {


  const offerDescription = await peerConn.createOffer();
  await peerConn.setLocalDescription(offerDescription);
  
  const offer = {
    sdp: offerDescription.sdp,
    type: offerDescription.type,
  };

  console.log(offer)

  if (conn !== null) {
    conn.send("request{ name: " + name + " sdp: " + offer);
  } 
} 


webcamButton.onclick = async () => {
  localStream = await navigator.mediaDevices.getUserMedia({video: true, audio: true});
  remoteStream = new MediaStream();

  localStream.getTracks().forEach((track) => {
    peerConn.addTrack(track, localStream);
  });

  peerConn.ontrack = event => {
    event.streams[0].getTracks().forEach(track => {
      remoteStream.addTrack(track);
    });
  };

  webcamVideo.srcObject = localStream;
  remoteVideo.srcObject = remoteStream;

};

